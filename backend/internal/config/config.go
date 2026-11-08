package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration values for the application.
type Config struct {
	// Database Connection Settings
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// Database Connection Pool Tuning
	DBMaxConns        int32
	DBMinConns        int32
	DBMaxConnLifetime int // in minutes
	DBMaxConnIdleTime int // in minutes

	// Application Server Settings
	ServerPort  string
	Environment string
	JWTSecret   string
}

// Load reads settings from .env file and environment variables.
func Load() (*Config, error) {
	// Try loading .env from current directory or parent directory (ignores error if file doesn't exist)
	_ = godotenv.Load(".env", "../.env")

	cfg := &Config{
		DBHost:            getEnv("DB_HOST", "localhost"),
		DBPort:            getEnv("DB_PORT", "5432"),
		DBUser:            getEnv("DB_USER", "findit_user"),
		DBPassword:        getEnv("DB_PASSWORD", "findit_secret"),
		DBName:            getEnv("DB_NAME", "findit_db"),
		DBSSLMode:         getEnv("DB_SSLMODE", "disable"),
		DBMaxConns:        getEnvAsInt32("DB_MAX_CONNS", 25),
		DBMinConns:        getEnvAsInt32("DB_MIN_CONNS", 5),
		DBMaxConnLifetime: getEnvAsInt("DB_MAX_CONN_LIFETIME_MINUTES", 60),
		DBMaxConnIdleTime: getEnvAsInt("DB_MAX_CONN_IDLE_MINUTES", 30),
		ServerPort:        getEnv("SERVER_PORT", "8080"),
		Environment:       getEnv("ENVIRONMENT", "development"),
		JWTSecret:         getEnv("JWT_SECRET", "default_secret_key_change_in_production"),
	}

	return cfg, nil
}

// DatabaseURL formats the complete PostgreSQL connection string (DSN).
func (c *Config) DatabaseURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser,
		c.DBPassword,
		c.DBHost,
		c.DBPort,
		c.DBName,
		c.DBSSLMode,
	)
}

// Helper: reads an environment variable with a string fallback default.
func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}

// Helper: reads an environment variable as an integer with a fallback.
func getEnvAsInt(key string, fallback int) int {
	valStr := getEnv(key, "")
	if val, err := strconv.Atoi(valStr); err == nil {
		return val
	}
	return fallback
}

// Helper: reads an environment variable as an int32 with a fallback.
func getEnvAsInt32(key string, fallback int32) int32 {
	valStr := getEnv(key, "")
	if val, err := strconv.ParseInt(valStr, 10, 32); err == nil {
		return int32(val)
	}
	return fallback
}
