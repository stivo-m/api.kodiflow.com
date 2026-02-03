package main

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stivo-m/api.kodiflow.com/internal/application/usecase"
	"github.com/stivo-m/api.kodiflow.com/internal/infrastructure/database"
	"github.com/stivo-m/api.kodiflow.com/pkg/config"
)

func main() {
	conStr := config.MustGetEnv("DATABASE_URL")

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, conStr)
	if err != nil {
		log.Fatalf("unable to create pgx pool on worker service: %v", err)
	}
	defer pool.Close()

	ctxPing, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(ctxPing); err != nil {
		log.Fatalf("unable to ping db from worker service: %v", err)
	}

	queries := database.New(pool)

	processor := &usecase.OutboxProcessor{
		Queries:       queries,
		Pool:          pool,
		PollInterval:  2 * time.Second,
		BatchSize:     10,
		MaxRetryCount: 5,
	}

	go processor.Start(ctx)

	// Block forever or until SIGTERM
	select {}
}
