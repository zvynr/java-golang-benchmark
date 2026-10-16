package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"benchmark-go/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresPool(cfg *config.Config) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to parse DSN: %w", err)
	}

	// Match Java HikariCP configuration for Apple-to-Apple benchmark
	poolConfig.MaxConns = cfg.DBMaxConns // 20 (HikariCP maximum-pool-size)
	poolConfig.MinConns = cfg.DBMinConns // 5 (HikariCP minimum-idle)

	if lifetime, err := time.ParseDuration(cfg.DBMaxConnLifetime); err == nil {
		poolConfig.MaxConnLifetime = lifetime // 30m (HikariCP max-lifetime)
	} else {
		poolConfig.MaxConnLifetime = 30 * time.Minute
	}

	if idleTime, err := time.ParseDuration(cfg.DBMaxConnIdleTime); err == nil {
		poolConfig.MaxConnIdleTime = idleTime // 10m (HikariCP idle-timeout)
	} else {
		poolConfig.MaxConnIdleTime = 10 * time.Minute
	}

	if timeout, err := time.ParseDuration(cfg.DBConnectTimeout); err == nil {
		poolConfig.ConnConfig.ConnectTimeout = timeout // 30s (HikariCP connection-timeout)
	} else {
		poolConfig.ConnConfig.ConnectTimeout = 30 * time.Second
	}

	poolConfig.HealthCheckPeriod = 1 * time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("[Database] Successfully connected via pgxpool (MaxConns: %d, MinConns: %d, MaxLifetime: %v, IdleTimeout: %v, ConnTimeout: %v)",
		cfg.DBMaxConns, cfg.DBMinConns, poolConfig.MaxConnLifetime, poolConfig.MaxConnIdleTime, poolConfig.ConnConfig.ConnectTimeout)

	return pool, nil
}
