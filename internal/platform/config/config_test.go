package config

import (
	"testing"
)

func setRequiredTestEnvironment(t *testing.T) {
	t.Helper()

	t.Setenv("POSTGRES_PASSWORD", "test-postgres-password")
	t.Setenv("MINIO_ROOT_PASSWORD", "test-minio-password")
}

func TestLoadUsesDefaults(t *testing.T) {
	t.Setenv("APP_PORT", "")
	t.Setenv("POSTGRES_PORT", "")
	t.Setenv("REDIS_HOST", "")
	t.Setenv("REDIS_PORT", "")
	t.Setenv("MINIO_USE_SSL", "")

	setRequiredTestEnvironment(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}

	if cfg.App.Environment != "development" {
		t.Errorf("expected development environment, got %q", cfg.App.Environment)
	}

	if cfg.App.Port != 8080 {
		t.Errorf("expected app port 8080, got %d", cfg.App.Port)
	}

	if cfg.Postgres.Port != 5434 {
		t.Errorf("expected PostgreSQL port 5434, got %d", cfg.Postgres.Port)
	}

	if cfg.Redis.Port != 6379 {
		t.Errorf("expected Redis port 6379, got %d", cfg.Redis.Port)
	}

	if cfg.MinIO.UseSSL {
		t.Error("expected MinIO SSL to default to false")
	}

	if cfg.Temporal.Address != "localhost:7233" {
		t.Errorf("unexpected Temporal address: %q", cfg.Temporal.Address)
	}
}

func TestLoadRequiresPostgresPassword(t *testing.T) {
	t.Setenv("POSTGRES_PASSWORD", "")
	t.Setenv("MINIO_ROOT_PASSWORD", "test-minio-password")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when PostgreSQL password is missing")
	}
}

func TestLoadRequiresMinIOPassword(t *testing.T) {
	t.Setenv("POSTGRES_PASSWORD", "test-postgres-password")
	t.Setenv("MINIO_ROOT_PASSWORD", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when MinIO password is missing")
	}
}

func TestLoadRejectsInvalidPort(t *testing.T) {
	setRequiredTestEnvironment(t)
	t.Setenv("APP_PORT", "70000")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid application port")
	}
}

func TestLoadReadsEnvironmentOverrides(t *testing.T) {
	setRequiredTestEnvironment(t)

	t.Setenv("APP_ENV", "test")
	t.Setenv("APP_PORT", "9090")
	t.Setenv("POSTGRES_HOST", "db.example.test")
	t.Setenv("MINIO_USE_SSL", "true")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}

	if cfg.App.Environment != "test" {
		t.Errorf("expected test environment, got %q", cfg.App.Environment)
	}

	if cfg.App.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.App.Port)
	}

	if cfg.Postgres.Host != "db.example.test" {
		t.Errorf("unexpected PostgreSQL host: %q", cfg.Postgres.Host)
	}

	if !cfg.MinIO.UseSSL {
		t.Error("expected MinIO SSL to be enabled")
	}
}

func TestLoadRejectsInvalidBoolean(t *testing.T) {
	setRequiredTestEnvironment(t)
	t.Setenv("MINIO_USE_SSL", "not-a-boolean")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid boolean")
	}
}
