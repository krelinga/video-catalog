package main

import (
	"context"
	"fmt"
	"log"

	"github.com/krelinga/video-catalog/internal"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func run() error {
	ctx := context.Background()
	
	// Load configuration
	cfg := internal.NewConfigFromEnv()

	// Create database pool
	pool, err := internal.NewDBPool(ctx, cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to create database pool: %w", err)
	}
	defer pool.Close()

	// Run migrations
	log.Println("Running database migrations...")
	if err := internal.MigrateUp(ctx, pool); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	log.Println("Migrations complete")

	return nil
}