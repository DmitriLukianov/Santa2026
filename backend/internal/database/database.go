package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PoolConfig struct {
	MaxConns int
	MinConns int
}

func NewDB(connString string, poolCfg PoolConfig) (*pgxpool.Pool, error) {
	ctx := context.Background()

	cfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	if poolCfg.MaxConns > 0 {
		cfg.MaxConns = int32(poolCfg.MaxConns)
	}
	if poolCfg.MinConns > 0 {
		cfg.MinConns = int32(poolCfg.MinConns)
	}
	cfg.MaxConnIdleTime = 5 * time.Minute
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.HealthCheckPeriod = 1 * time.Minute

	dbpool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	if err := dbpool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	return dbpool, nil
}
