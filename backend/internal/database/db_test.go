package database

import (
	"context"
	"testing"
	"time"

	"github.com/MohammedaminSa/FindIt/backend/internal/config"
)

func TestNewPoolConfig(t *testing.T) {
	cfg := &config.Config{
		DBHost:            "localhost",
		DBPort:            "5432",
		DBUser:            "findit_user",
		DBPassword:        "findit_secret",
		DBName:            "findit_db",
		DBSSLMode:         "disable",
		DBMaxConns:        30,
		DBMinConns:        8,
		DBMaxConnLifetime: 45,
		DBMaxConnIdleTime: 15,
	}

	poolConfig, err := NewPoolConfig(cfg)
	if err != nil {
		t.Fatalf("NewPoolConfig failed unexpectedly: %v", err)
	}

	if poolConfig.MaxConns != 30 {
		t.Errorf("expected MaxConns=30, got %d", poolConfig.MaxConns)
	}

	if poolConfig.MinConns != 8 {
		t.Errorf("expected MinConns=8, got %d", poolConfig.MinConns)
	}

	expectedLifetime := 45 * time.Minute
	if poolConfig.MaxConnLifetime != expectedLifetime {
		t.Errorf("expected MaxConnLifetime=%v, got %v", expectedLifetime, poolConfig.MaxConnLifetime)
	}

	expectedIdleTime := 15 * time.Minute
	if poolConfig.MaxConnIdleTime != expectedIdleTime {
		t.Errorf("expected MaxConnIdleTime=%v, got %v", expectedIdleTime, poolConfig.MaxConnIdleTime)
	}

	expectedHealthCheck := 1 * time.Minute
	if poolConfig.HealthCheckPeriod != expectedHealthCheck {
		t.Errorf("expected HealthCheckPeriod=%v, got %v", expectedHealthCheck, poolConfig.HealthCheckPeriod)
	}
}

func TestNewPoolConfigInvalidDSN(t *testing.T) {
	cfg := &config.Config{
		DBHost: "invalid host with spaces",
		DBPort: "invalid-port",
	}

	_, err := NewPoolConfig(cfg)
	if err == nil {
		t.Error("expected error for invalid database configuration, got nil")
	}
}

func TestNewPoolUnreachableHost(t *testing.T) {
	cfg := &config.Config{
		DBHost:            "127.0.0.1",
		DBPort:            "54329", // unreachable port
		DBUser:            "nonexistent_user",
		DBPassword:        "nonexistent_password",
		DBName:            "nonexistent_db",
		DBSSLMode:         "disable",
		DBMaxConns:        2,
		DBMinConns:        1,
		DBMaxConnLifetime: 5,
		DBMaxConnIdleTime: 2,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	pool, err := NewPool(ctx, cfg)
	if err == nil {
		if pool != nil {
			pool.Close()
		}
		t.Error("expected NewPool to return an error when target is unreachable, got nil")
	}
}
