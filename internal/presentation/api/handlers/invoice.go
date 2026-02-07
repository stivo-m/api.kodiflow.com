package handlers

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stivo-m/api.kodiflow.com/internal/application/dto"
	"github.com/stivo-m/api.kodiflow.com/internal/application/usecase"
	"github.com/stivo-m/api.kodiflow.com/internal/infrastructure/database"
	"github.com/stivo-m/api.kodiflow.com/internal/presentation/api/middleware"
	"github.com/stivo-m/api.kodiflow.com/pkg/api"
	"github.com/stivo-m/api.kodiflow.com/pkg/helpers"
)

type invoiceHandler struct {
	queries *database.Queries
	usecase *usecase.InvoiceUsecase
}

// new invoice handler
func newInvoiceHandler(queries *database.Queries, pool *pgxpool.Pool) *invoiceHandler {
	return &invoiceHandler{
		queries: queries,
		usecase: usecase.NewInvoiceUsecase(queries, pool),
	}
}

// Register routes
func (h *invoiceHandler) RegisterRoutes(router *http.ServeMux) {
	r := http.NewServeMux()

	// Invoices
	r.Handle("POST /", helpers.ValidateBody(h.createInvoiceHandler))
	r.HandleFunc("GET /", h.listInvoicesHandler)

	protected := middleware.CreateMiddlewareStack(
		middleware.AuthMiddleware,
		middleware.BusinessContextMiddleware(h.queries),
	)

	router.Handle("/invoices/", http.StripPrefix("/invoices", protected(r)))
}

// creates a new invoice
func (h *invoiceHandler) createInvoiceHandler(w http.ResponseWriter, r *http.Request, payload dto.CreateInvoiceDto) {
	result := h.usecase.CreateInvoiceUsecase(r.Context(), &payload)
	api.ProcessUsecaseResponse(result, w)
}

// List invoices for business
func (h *invoiceHandler) listInvoicesHandler(w http.ResponseWriter, r *http.Request) {
	result := h.usecase.ListInvociesUsecase(r.Context())
	api.ProcessUsecaseResponse(result, w)
}
