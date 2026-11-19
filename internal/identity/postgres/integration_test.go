//go:build integration

package postgres

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	integrationPool *pgxpool.Pool
	integrationRepo *ActionTokenRepository
)

func TestMain(m *testing.M) {
	if err := setupIntegrationDatabase(); err != nil {
		fmt.Fprintln(os.Stderr, "integration test setup failed:", err)
		os.Exit(1)
	}

	exitCode := m.Run()

	integrationPool.Close()
	os.Exit(exitCode)
}

func setupIntegrationDatabase() error {
	// Fail closed: integration tests must never touch the development database.
	if os.Getenv("POSTGRES_DB") != "hireos_test" {
		return fmt.Errorf(
			"integration tests require POSTGRES_DB=hireos_test",
		)
	}

	host := os.Getenv("POSTGRES_HOST")
	portValue := os.Getenv("POSTGRES_PORT")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")

	if host == "" || portValue == "" || user == "" || password == "" {
		return fmt.Errorf("required PostgreSQL environment variables are missing")
	}

	port, err := strconv.Atoi(portValue)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("POSTGRES_PORT must be a valid TCP port")
	}

	connURL := url.URL{
		Scheme: "postgres",
		User:   url.User(user),
		Host:   net.JoinHostPort(host, strconv.Itoa(port)),
		Path:   "/hireos_test",
	}

	poolConfig, err := pgxpool.ParseConfig(connURL.String())
	if err != nil {
		return fmt.Errorf("parse test database configuration: %w", err)
	}

	poolConfig.ConnConfig.Password = password
	poolConfig.MaxConns = 10

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return fmt.Errorf("create test database pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return fmt.Errorf("connect to test database: %w", err)
	}

	var databaseName string
	if err := pool.QueryRow(
		ctx,
		`SELECT current_database()`,
	).Scan(&databaseName); err != nil {
		pool.Close()
		return fmt.Errorf("verify test database name: %w", err)
	}

	if databaseName != "hireos_test" {
		pool.Close()
		return fmt.Errorf("refusing to test against unexpected database")
	}

	var version int64
	var dirty bool

	err = pool.QueryRow(
		ctx,
		`SELECT version, dirty FROM schema_migrations LIMIT 1`,
	).Scan(&version, &dirty)
	if err != nil {
		pool.Close()
		return fmt.Errorf("read database migration version: %w", err)
	}

	if version != 3 || dirty {
		pool.Close()
		return fmt.Errorf(
			"test database must be at migration version 3 with dirty=false",
		)
	}

	integrationPool = pool
	integrationRepo = NewActionTokenRepository(pool)

	return nil
}
