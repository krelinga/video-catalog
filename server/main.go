package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/krelinga/video-catalog/internal"
	"github.com/krelinga/video-catalog/vcrest"
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

	// Create server instance
	srv := &Server{
		Config: cfg,
	}
	strictHandler := vcrest.NewStrictHandler(srv, nil)
	httpHandler := vcrest.Handler(strictHandler)

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.ServerPort),
		Handler: httpHandler,
	}

	// Start HTTP server
	log.Println("Starting HTTP server on port", cfg.ServerPort)
	if err := httpServer.ListenAndServe(); err != nil {
		return fmt.Errorf("server failed: %w", err)
	}

	return nil
}
