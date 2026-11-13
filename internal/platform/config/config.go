package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	App      AppConfig
	Postgres PostgresConfig
	Redis    RedisConfig
	MinIO    MinIOConfig
	Temporal TemporalConfig
}

type AppConfig struct {
	Environment string
	Port        int
}

type PostgresConfig struct {
	Host     string
	Port     int
	Database string
	User     string
	Password string
}

type RedisConfig struct {
	Host string
	Port int
}

type MinIOConfig struct {
	RootUser     string
	RootPassword string
	Endpoint     string
	Bucket       string
	UseSSL       bool
}

type TemporalConfig struct {
	Address   string
	NameSpace string
}

func Load() (Config, error) {
	cfg := Config{
		App: AppConfig{
			Environment: envOrDefault("APP_ENV", "development"),
			Port:        8080,
		},
		Postgres: PostgresConfig{
			Host:     envOrDefault("POSTGRES_HOST", "localhost"),
			Port:     5434,
			Database: envOrDefault("POSTGRES_DB", "hireos"),
			User:     envOrDefault("POSTGRES_USER", "hireos"),
		},
		Redis: RedisConfig{
			Host: "localhost",
			Port: 6379,
		},
		MinIO: MinIOConfig{
			RootUser: envOrDefault("MINIO_ROOT_USER", "hireos_admin"),
			Endpoint: envOrDefault("MINIO_ENDPOINT", "localhost:9000"),
			Bucket:   envOrDefault("MINIO_BUCKET", "hireos-local"),
		},
		Temporal: TemporalConfig{
			Address:   envOrDefault("TEMPORAL_ADDRESS", "localhost:7233"),
			NameSpace: envOrDefault("TEMPORAL_NAMESPACE", "default"),
		},
	}

	var err error

	if cfg.App.Port, err = portFromEnv("APP_PORT", cfg.App.Port); err != nil {
		return Config{}, err
	}

	if cfg.Postgres.Port, err = portFromEnv("POSTGRES_PORT", cfg.Postgres.Port); err != nil {
		return Config{}, err
	}

	if cfg.Redis.Host, err = optionalEnv("REDIS_HOST", cfg.Redis.Host); err != nil {
		return Config{}, err
	}

	if cfg.Redis.Port, err = portFromEnv("REDIS_PORT", cfg.Redis.Port); err != nil {
		return Config{}, err
	}

	if cfg.Postgres.Password, err = requiredEnv("POSTGRES_PASSWORD"); err != nil {
		return Config{}, err
	}

	if cfg.MinIO.RootPassword, err = requiredEnv("MINIO_ROOT_PASSWORD"); err != nil {
		return Config{}, err
	}

	if cfg.MinIO.UseSSL, err = boolFromEnv("MINIO_USE_SSL", false); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func envOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	return value
}

func optionalEnv(key, defaultValue string) (string, error) {
	return envOrDefault(key, defaultValue), nil
}

func requiredEnv(key string) (string, error) {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		return "", fmt.Errorf("%s is required.", key)
	}

	return value, nil
}

func portFromEnv(key string, defaultValue int) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue, nil
	}

	port, err := strconv.Atoi(value)
	if err != nil || port < 1 || port > 65535 {
		return 0, fmt.Errorf("%s must be a valid TCP port (1-65535)", key)
	}

	return port, nil
}

func boolFromEnv(key string, defaultValue bool) (bool, error) {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue, nil
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("%s must be a boolean", key)
	}

	return parsed, nil
}
