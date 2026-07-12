package db

import (
    "context"
    "fmt"
    "os"

    "github.com/jackc/pgx/v5/pgxpool"
)

// NewPool creates a connection pool and verifies the connection.
// It returns an error if the database is unreachable — we want to
// fail fast on startup rather than serve requests with no DB.
func NewPool(ctx context.Context) (*pgxpool.Pool, error) {
    url := os.Getenv("DATABASE_URL")
    if url == "" {
        return nil, fmt.Errorf("DATABASE_URL environment variable not set")
    }

    pool, err := pgxpool.New(ctx, url)
    if err != nil {
        return nil, fmt.Errorf("unable to create connection pool: %w", err)
    }

    // Ping verifies at least one connection can be established.
    // If this fails, we know immediately — not on the first request.
    if err := pool.Ping(ctx); err != nil {
        return nil, fmt.Errorf("unable to reach database: %w", err)
    }

    return pool, nil
}