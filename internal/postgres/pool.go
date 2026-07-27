package postgres

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PoolConfig struct {
	MaxConns          int32
	MinConns          int32
	MaxConnIdleTime   time.Duration
	MaxConnLifetime   time.Duration
	HealthCheckPeriod time.Duration
}

func NewPool(dbUrl string, cfg PoolConfig) (*pgxpool.Pool, error) {
	// Create config for pool
	config, err := pgxpool.ParseConfig(dbUrl)
	if err != nil {
		return nil, fmt.Errorf("config parsing failed: %w", err)
	}

	config.MaxConnIdleTime = cfg.MaxConnIdleTime
	config.MaxConnLifetime = cfg.MaxConnLifetime
	config.MaxConns = cfg.MaxConns
	config.MinConns = cfg.MinConns
	config.HealthCheckPeriod = cfg.HealthCheckPeriod

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Creating pool with config
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("pool creation failed: %w", err)
	}

	// Verify connection with ping
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping failed: %w", err)
	}

	return pool, nil
}

func GetUsersPostgresDSN() string {
	dbUser := os.Getenv("POSTGRES_USERS_USER")
	dbPassword := os.Getenv("POSTGRES_USERS_PASSWORD")
	dbHost := os.Getenv("POSTGRES_USERS_HOST")
	dbPort := os.Getenv("POSTGRES_USERS_PORT")
	dbName := os.Getenv("POSTGRES_USERS_DB")

	return fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName)
}

func GetPinsPostgresDSN() string {
	dbUser := os.Getenv("POSTGRES_PINS_USER")
	dbPassword := os.Getenv("POSTGRES_PINS_PASSWORD")
	dbHost := os.Getenv("POSTGRES_PINS_HOST")
	dbPort := os.Getenv("POSTGRES_PINS_PORT")
	dbName := os.Getenv("POSTGRES_PINS_DB")

	return fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName)
}
