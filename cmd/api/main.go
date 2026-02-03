package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"github.com/stivo-m/api.kodiflow.com/internal/infrastructure/database"
	"github.com/stivo-m/api.kodiflow.com/internal/presentation/api/server"
	"github.com/stivo-m/api.kodiflow.com/pkg/config"
)

func main() {
	conStr := config.MustGetEnv("DATABASE_URL")

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, conStr)
	if err != nil {
		log.Fatalf("unable to create pgx pool: %v", err)
	}
	defer pool.Close()

	ctxPing, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(ctxPing); err != nil {
		log.Fatalf("unable to ping db: %v", err)
	}

	queries := database.New(pool)
	config := server.HttpServerConfig{
		Address: ":8080",
		Queries: queries,
		Pool:    pool,
	}

	svr := server.NewHttpServer(config)
	router := svr.MountRoutes()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := svr.RunServer(router); err != nil {
			log.Printf("Http server failed to start, error: %v", err)
		}
	}()

	sig := <-quit
	log.Printf("Received a signal to shutdown the server, signal:  %v", sig)

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := svr.ShutdownServer(ctx); err != nil {
		log.Printf("Failed to shutdown the server, error: %v", err)
	} else {
		log.Println("Server shutdown successfully")
	}
}
