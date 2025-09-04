package db

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

// InitDB connects to Neon using DATABASE_URL and sets Pool.
// call db.InitDB(context.Background()) from main()
func InitDB(ctx context.Context) (*pgxpool.Pool, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not set")
	}

	// optional: parse config if you want to tune pool
	cfg, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		log.Fatalf("pgx ParseConfig error: %v", err)
		return nil, err
	}
	cfg.MaxConns = 10

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
		return nil, err
	}

	// test connection
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("failed to ping db: %v", err)
		return nil, err
	}

	Pool = pool
	return pool, nil
}
