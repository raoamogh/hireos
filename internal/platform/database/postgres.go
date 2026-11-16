package database

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/raoamogh/hireos/internal/platform/config"
)

func NewPool(ctx context.Context, cfg config.PostgresConfig) (*pgxpool.Pool, error) {
	poolConfig, err := buildPoolConfig(cfg)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create PostgreSQL connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("verify PostgreSQL connectivity: %w", err)
	}

	return pool, nil
}

func buildPoolConfig(cfg config.PostgresConfig) (*pgxpool.Config, error) {
	if cfg.Host == "" {
		return nil, fmt.Errorf("PostgreSQL host is required")
	}

	if cfg.Port < 1 || cfg.Port > 65535 {
		return nil, fmt.Errorf("PostgreSQL port must be between 1 and 65535")
	}

	if cfg.Database == "" {
		return nil, fmt.Errorf("PostgreSQL database name is required")
	}

	if cfg.User == "" {
		return nil, fmt.Errorf("PostgreSQL username is required")
	}

	if cfg.Password == "" {
		return nil, fmt.Errorf("PostgreSQL password is required")
	}

	connURL := url.URL{
		Scheme: "postgres",
		User:   url.User(cfg.User),
		Host:   net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)),
		Path:   "/" + cfg.Database,
	}

	poolConfig, err := pgxpool.ParseConfig(connURL.String())
	if err != nil {
		return nil, fmt.Errorf("parse PostgreSQL pool configuration: %w", err)
	}

	poolConfig.ConnConfig.Password = cfg.Password

	return poolConfig, nil
}
