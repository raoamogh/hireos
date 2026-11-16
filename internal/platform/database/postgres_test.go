package database

import (
	"testing"

	"github.com/raoamogh/hireos/internal/platform/config"
)

func TestBuildPoolConfig(t *testing.T) {
	cfg := config.PostgresConfig{
		Host:     "localhost",
		Port:     5434,
		Database: "hireos",
		User:     "hireos",
		Password: "test-password",
	}

	poolConfig, err := buildPoolConfig(cfg)
	if err != nil {
		t.Fatalf("buildPoolConfig() returned an error: %v", err)
	}

	if poolConfig.ConnConfig.Host != cfg.Host {
		t.Errorf("expected host %q, got %q",
			cfg.Host, poolConfig.ConnConfig.Host)
	}

	if poolConfig.ConnConfig.Port != uint16(cfg.Port) {
		t.Errorf("expected port %d, got %d",
			cfg.Port, poolConfig.ConnConfig.Port)
	}

	if poolConfig.ConnConfig.Database != cfg.Database {
		t.Errorf("expected database %q, got %q",
			cfg.Database, poolConfig.ConnConfig.Database)
	}

	if poolConfig.ConnConfig.User != cfg.User {
		t.Errorf("expected user %q, got %q",
			cfg.User, poolConfig.ConnConfig.User)
	}

	if poolConfig.ConnConfig.Password != cfg.Password {
		t.Error("PostgreSQL password was not configured correctly")
	}

	if poolConfig.MaxConns < 1 {
		t.Error("expected at least one maximum connection")
	}

	if poolConfig.ConnConfig.Host != "localhost" {
		t.Errorf("expected connection host localhost, got %q",
			poolConfig.ConnConfig.Host)
	}

	if poolConfig.ConnConfig.Port != 5434 {
		t.Errorf("expected connection port 5434, got %d",
			poolConfig.ConnConfig.Port)
	}

	for _, fallback := range poolConfig.ConnConfig.Fallbacks {
		if fallback.Host != "localhost" || fallback.Port != 5434 {
			t.Errorf(
				"unexpected fallback address: %s:%d",
				fallback.Host,
				fallback.Port,
			)
		}
	}
}

func TestBuildPoolConfigRejectsInvalidSettings(t *testing.T) {
	tests := []struct {
		name   string
		config config.PostgresConfig
	}{
		{
			name: "missing host",
			config: config.PostgresConfig{
				Port: 5434, Database: "hireos",
				User: "hireos", Password: "test",
			},
		},
		{
			name: "invalid port",
			config: config.PostgresConfig{
				Host: "localhost", Port: 70000,
				Database: "hireos", User: "hireos",
				Password: "test",
			},
		},
		{
			name: "missing database",
			config: config.PostgresConfig{
				Host: "localhost", Port: 5434,
				User: "hireos", Password: "test",
			},
		},
		{
			name: "missing username",
			config: config.PostgresConfig{
				Host: "localhost", Port: 5434,
				Database: "hireos", Password: "test",
			},
		},
		{
			name: "missing password",
			config: config.PostgresConfig{
				Host: "localhost", Port: 5434,
				Database: "hireos", User: "hireos",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := buildPoolConfig(tt.config)
			if err == nil {
				t.Fatal("expected an error, got nil")
			}
		})
	}

}
