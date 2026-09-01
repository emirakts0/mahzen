package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/emirakts0/mahzen/internal/config"
)

// NewPool creates a pgxpool from config. Zero-valued pool settings keep the
// pgx defaults instead of panicking on invalid values.
func NewPool(ctx context.Context, cfg config.DatabaseConfig) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("parsing database config: %w", err)
	}

	if cfg.Pool.MaxConns > 0 {
		poolCfg.MaxConns = int32(cfg.Pool.MaxConns)
	}
	if cfg.Pool.MinConns > 0 {
		poolCfg.MinConns = int32(cfg.Pool.MinConns)
	}
	if cfg.Pool.MaxConnLifetime > 0 {
		poolCfg.MaxConnLifetime = cfg.Pool.MaxConnLifetime
	}
	if cfg.Pool.MaxConnIdleTime > 0 {
		poolCfg.MaxConnIdleTime = cfg.Pool.MaxConnIdleTime
	}
	if cfg.Pool.HealthCheckPeriod > 0 {
		poolCfg.HealthCheckPeriod = cfg.Pool.HealthCheckPeriod
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("creating connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	return pool, nil
}
