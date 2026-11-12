package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type databaseConfig struct {
	host     string
	port     string
	name     string
	user     string
	password string
}

func main() {
	if err := run(); err != nil {
		slog.Error("migration command failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	command := "up"

	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	switch command {
	case "up", "down", "version":
	default:
		return fmt.Errorf(
			"unknown command %q; supported commands: up, down, version",
			command,
		)
	}

	cfg, err := loadDatabaseConfig()
	if err != nil {
		return err
	}

	workingDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	migrationsPath := filepath.Join(workingDir, "db", "migrations")

	sourceURL := (&url.URL{
		Scheme: "file",
		Path:   filepath.ToSlash(migrationsPath),
	}).String()

	databaseURL := (&url.URL{
		Scheme: "pgx5",
		User:   url.UserPassword(cfg.user, cfg.password),
		Host:   net.JoinHostPort(cfg.host, cfg.port),
		Path:   "/" + cfg.name,
	}).String()

	m, err := migrate.New(sourceURL, databaseURL)
	if err != nil {
		return fmt.Errorf("initialize migration engine: %w", err)
	}

	defer func() {
		sourceErr, databaseErr := m.Close()

		if sourceErr != nil {
			slog.Warn("close migration source", "error", sourceErr)
		}

		if databaseErr != nil {
			slog.Warn("close migration database", "error", databaseErr)
		}
	}()

	switch command {
	case "up":
		err = m.Up()

	case "down":
		// Roll back exactly one migration.
		err = m.Steps(-1)

	case "version":
		version, dirty, versionErr := m.Version()
		if errors.Is(versionErr, migrate.ErrNilVersion) {
			slog.Info("no migrations have been applied")
			return nil
		}
		if versionErr != nil {
			return fmt.Errorf("read migration version: %w", versionErr)
		}

		slog.Info(
			"database migration version",
			"version", version,
			"dirty", dirty,
		)
		return nil
	}

	if errors.Is(err, migrate.ErrNoChange) {
		slog.Info("database is already at the requested migration state")
		return nil
	}

	if err != nil {
		return fmt.Errorf("migration command %q failed: %w", command, err)
	}

	slog.Info("database migration command completed", "command", command)
	return nil
}

func loadDatabaseConfig() (databaseConfig, error) {
	requiredVariables := []string{
		"POSTGRES_HOST",
		"POSTGRES_PORT",
		"POSTGRES_DB",
		"POSTGRES_USER",
		"POSTGRES_PASSWORD",
	}

	values := make(map[string]string, len(requiredVariables))

	for _, key := range requiredVariables {
		value := os.Getenv(key)
		if value == "" {
			return databaseConfig{}, fmt.Errorf(
				"required environment variable %s is not set",
				key,
			)
		}

		values[key] = value
	}

	return databaseConfig{
		host:     values["POSTGRES_HOST"],
		port:     values["POSTGRES_PORT"],
		name:     values["POSTGRES_DB"],
		user:     values["POSTGRES_USER"],
		password: values["POSTGRES_PASSWORD"],
	}, nil
}
