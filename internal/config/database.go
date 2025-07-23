package config

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InitPostgresPool(connectionString string) (*pgxpool.Pool, error) {
	// Parse configuration
	config, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DB config: %w", err)
	}

	// Configure connection pool (optimized for payments)
	config.MaxConns = 25                  // Default is usually too high
	config.MinConns = 2                   // Keep some warm connections
	config.MaxConnLifetime = 1 * time.Hour // Refresh connections periodically
	config.HealthCheckPeriod = 1 * time.Minute

	// Create the pool
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	return pool, nil
}