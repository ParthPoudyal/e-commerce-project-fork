package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(databaseURL string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database URL: %w", err)
	}

	config.MaxConns = 20
	config.MinConns = 2
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnIdleTime = 5 * time.Minute

	var pool *pgxpool.Pool
	var lastErr error

	for attempts := 1; attempts <= 5; attempts++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		pool, err = pgxpool.NewWithConfig(ctx, config)
		if err == nil {
			if err = pool.Ping(ctx); err == nil {
				cancel()
				log.Println("Connected to PostgreSQL")
				return pool, nil
			}
			pool.Close()
		}
		cancel()
		lastErr = err
		log.Printf("Waiting for database connection (attempt %d/5)...", attempts)
		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("failed to connect to database after 5 attempts: %w", lastErr)
}
