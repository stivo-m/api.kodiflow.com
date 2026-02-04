package handlers

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stivo-m/api.kodiflow.com/internal/infrastructure/database"
)

type apiHandlers struct {
	router  *http.ServeMux
	queries *database.Queries
	pool    *pgxpool.Pool
}

// Instantiate a new api handler
func NewApiHandlers(
	router *http.ServeMux,
	queries *database.Queries,
	pool *pgxpool.Pool,
) *apiHandlers {
	return &apiHandlers{router: router, queries: queries, pool: pool}
}

// Register associated routes
func (h *apiHandlers) RegisterRoutes() {
	h.router.HandleFunc("/health", healthCheckHandler)

	// // Waitlists
	// wHandler := newWaitlistHandler(h.queries, h.pool)
	// wHandler.RegisterRoutes(h.router)
	//
}
