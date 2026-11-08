package database

import (
	"context"
	"fmt"
	"time"

	"github.com/MohammedaminSa/FindIt/backend/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPoolConfig constructs and tunes the pgxpool.Config from the application Config.
func NewPoolConfig(cfg *config.Config) (*pgxpool.Config, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL())
	if err != nil {
		return nil, fmt.Errorf("failed to parse database connection URL: %w", err)
	}

	// Apply connection pool tuning parameters
	poolConfig.MaxConns = cfg.DBMaxConns
	poolConfig.MinConns = cfg.DBMinConns
	poolConfig.MaxConnLifetime = time.Duration(cfg.DBMaxConnLifetime) * time.Minute
	poolConfig.MaxConnIdleTime = time.Duration(cfg.DBMaxConnIdleTime) * time.Minute
	poolConfig.HealthCheckPeriod = 1 * time.Minute

	return poolConfig, nil
}

// NewPool initializes and validates a high-performance PostgreSQL connection pool.
// It immediately pings the database to ensure network connectivity and credentials are valid.
func NewPool(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	poolConfig, err := NewPoolConfig(cfg)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize pgx connection pool: %w", err)
	}

	// Verify the database is reachable with a dedicated timeout
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	return pool, nil
}
