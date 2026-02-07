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

type businessHandler struct {
	queries *database.Queries
	usecase *usecase.BusinessUsecase
}

// New business handler
func newBusinessHandler(queries *database.Queries, pool *pgxpool.Pool) *businessHandler {
	return &businessHandler{
		queries: queries,
		usecase: usecase.NewBusinessUsecase(queries, pool),
	}
}

// Register routes
func (h *businessHandler) RegisterRoutes(router *http.ServeMux) {
	r := http.NewServeMux()
	r.Handle("POST /", helpers.ValidateBody(h.createBusinessHandler))
	r.HandleFunc("GET /", h.listBusinessessHandler)
	protected := middleware.CreateMiddlewareStack(
		middleware.AuthMiddleware,
	)

	router.Handle("/businesses/", http.StripPrefix("/businesses", protected(r)))

	// Verify business ownership
	br := http.NewServeMux()
	br.Handle("POST /users", helpers.ValidateBody(h.addUsersToBusinessHandler))
	br.Handle("POST /kyc", helpers.ValidateBody(h.createBusinessKycHandler))
	br.HandleFunc("GET /kyc", h.listKycForBusiness)

	businessProtected := middleware.CreateMiddlewareStack(
		middleware.AuthMiddleware,
		middleware.BusinessContextMiddleware(h.queries),
	)
	router.Handle("/business/", http.StripPrefix("/business", businessProtected(br)))
}

// Creates a new business
func (h *businessHandler) createBusinessHandler(w http.ResponseWriter, r *http.Request, payload dto.CreateBusinessDto) {
	result := h.usecase.CreateBusinessUsecase(r.Context(), &payload)
	api.ProcessUsecaseResponse(result, w)
}

// Adds users to a business
func (h *businessHandler) addUsersToBusinessHandler(w http.ResponseWriter, r *http.Request, payload dto.AddUserToBusinessDto) {
	result := h.usecase.AddUserToBusinessUsecase(r.Context(), &payload)
	api.ProcessUsecaseResponse(result, w)
}

// Find all businessess belonging to user
func (h *businessHandler) listBusinessessHandler(w http.ResponseWriter, r *http.Request) {
	result := h.usecase.ListBusinessessForUserUsecase(r.Context())
	api.ProcessUsecaseResponse(result, w)
}

// Add business kyc
func (h *businessHandler) createBusinessKycHandler(w http.ResponseWriter, r *http.Request, payload dto.CreateBusinessKycDto) {
	businessId, err := helpers.BusinessFromContext(r.Context())
	if err != nil {
		err := api.ApiResponse{
			Code:    400,
			Message: "Invalid business id",
		}
		api.ProcessUsecaseResponse(&err, w)
		return
	}

	result := h.usecase.CreateBusinessKycUsecase(r.Context(), businessId, &payload)
	api.ProcessUsecaseResponse(result, w)
}

// List kyc for business
func (h *businessHandler) listKycForBusiness(w http.ResponseWriter, r *http.Request) {
	businessId, err := helpers.BusinessFromContext(r.Context())
	if err != nil {
		err := api.ApiResponse{
			Code:    400,
			Message: "Invalid business id",
		}
		api.ProcessUsecaseResponse(&err, w)
		return
	}

	result := h.usecase.ListKycForBusinessUsecase(r.Context(), businessId)
	api.ProcessUsecaseResponse(result, w)
}
