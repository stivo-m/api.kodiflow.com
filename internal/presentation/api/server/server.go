package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stivo-m/api.kodiflow.com/internal/infrastructure/database"
	"github.com/stivo-m/api.kodiflow.com/internal/presentation/api/handlers"
	"github.com/stivo-m/api.kodiflow.com/internal/presentation/api/middleware"
)

type HttpServerConfig struct {
	Address string
	Queries *database.Queries
	Pool    *pgxpool.Pool
}

type httpServer struct {
	Config HttpServerConfig
	Svr    *http.Server
}

// Instantiates a new http server
func NewHttpServer(config HttpServerConfig) *httpServer {
	return &httpServer{Config: config}
}

// Setup all the necessary routes within the server
func (s *httpServer) MountRoutes() http.Handler {
	router := http.NewServeMux()

	v1 := http.NewServeMux()

	router.Handle("/api/v1/", http.StripPrefix("/api/v1", v1))

	apiHandlers := handlers.NewApiHandlers(v1, s.Config.Queries, s.Config.Pool)
	apiHandlers.RegisterRoutes()

	return router
}

// Run the server with the current configurations
func (s *httpServer) RunServer(router http.Handler) error {
	stack := middleware.CreateMiddlewareStack()

	s.Svr = &http.Server{
		Addr:         s.Config.Address,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
		IdleTimeout:  time.Minute,
		Handler:      stack(router),
	}

	slog.Info("Database connection successful")
	slog.Info(fmt.Sprintf("Server started at http://localhost%s", s.Config.Address))
	return s.Svr.ListenAndServe()
}

// Gracefully shutdown the server after completing all pending requests
func (s *httpServer) ShutdownServer(ctx context.Context) error {
	if s.Svr != nil {
		slog.Info("Shutting down the server...")
		return s.Svr.Shutdown(ctx)
	}

	return nil
}
