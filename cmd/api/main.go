package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"github.com/stivo-m/api.kodiflow.com/internal/infrastructure/database"
	"github.com/stivo-m/api.kodiflow.com/internal/presentation/api/middleware"
	"github.com/stivo-m/api.kodiflow.com/internal/presentation/api/server"
	"github.com/stivo-m/api.kodiflow.com/pkg/config"
)

func main() {

	// Setup logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	slog.SetDefault(logger)

	// Setup database connection
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

	// Setup the server
	queries := database.New(pool)
	config := server.HttpServerConfig{
		Address: ":8080",
		Queries: queries,
		Pool:    pool,
	}

	svr := server.NewHttpServer(config)
	router := svr.MountRoutes()

	stack := middleware.CreateMiddlewareStack(
		middleware.RequestLogger,
	)

	router = stack(router)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := svr.RunServer(router); err != nil {
			slog.Error("Http server failed to start", "error", err)
		}
	}()

	sig := <-quit
	slog.Info("Received a signal to shutdown the server", "signal", sig)

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := svr.ShutdownServer(ctx); err != nil {
		slog.Error("Failed to shutdown the server", "error", err)
	} else {
		slog.Info("Server shutdown successfully")
	}
}
