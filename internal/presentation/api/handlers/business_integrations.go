package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stivo-m/api.kodiflow.com/internal/application/dto"
	"github.com/stivo-m/api.kodiflow.com/internal/application/usecase"
	"github.com/stivo-m/api.kodiflow.com/internal/infrastructure/database"
	"github.com/stivo-m/api.kodiflow.com/internal/presentation/api/middleware"
	"github.com/stivo-m/api.kodiflow.com/pkg/api"
	"github.com/stivo-m/api.kodiflow.com/pkg/helpers"
)

type integrationsHandler struct {
	queries *database.Queries
	usecase *usecase.BusinessIntegrationUsecase
}

// new integrations handler
func newIntegrationsHandler(queries *database.Queries, pool *pgxpool.Pool) *integrationsHandler {
	return &integrationsHandler{
		queries: queries,
		usecase: usecase.NewBusinessIntegrationUsecase(queries, pool),
	}
}

// Register routes
func (h *integrationsHandler) RegisterRoutes(router *http.ServeMux) {
	r := http.NewServeMux()

	// Integrations
	r.Handle("POST /", helpers.ValidateBody(h.createIntegrationHandler))
	r.HandleFunc("GET /", h.listIntegrationsHandler)
	r.HandleFunc("DELETE /{integrationId}", h.deleteIntegrationHandler)

	// Secrets
	r.Handle("POST /{integrationId}/secrets", helpers.ValidateBody(h.createIntegrationSecretHandler))
	r.Handle("DELETE /{integrationId}/secrets", helpers.ValidateBody(h.deleteIntegrationSecretHandler))

	protected := middleware.CreateMiddlewareStack(
		middleware.AuthMiddleware,
		middleware.BusinessContextMiddleware(h.queries),
	)

	router.Handle("/integrations/", http.StripPrefix("/integrations", protected(r)))
}

// creates a new integration
func (h *integrationsHandler) createIntegrationHandler(w http.ResponseWriter, r *http.Request, payload dto.CreateIntegrationDto) {
	result := h.usecase.CreateIntegrationUsecase(r.Context(), &payload)
	api.ProcessUsecaseResponse(result, w)
}

// List integrations for business
func (h *integrationsHandler) listIntegrationsHandler(w http.ResponseWriter, r *http.Request) {
	result := h.usecase.ListIntegrationsUsecase(r.Context())
	api.ProcessUsecaseResponse(result, w)
}

// Deletes an integration
func (h *integrationsHandler) deleteIntegrationHandler(w http.ResponseWriter, r *http.Request) {
	integrationIdStr := r.PathValue("integrationId")
	integrationId, err := uuid.Parse(integrationIdStr)
	if err != nil {
		res := api.ApiResponse{
			Code:    400,
			Errors:  err,
			Message: "Malformed integration id",
		}
		api.ProcessUsecaseResponse(&res, w)
		return
	}
	result := h.usecase.DeleteBusinessIntegrationUsecase(r.Context(), integrationId)
	api.ProcessUsecaseResponse(result, w)
}

// Creates an integration secret
func (h *integrationsHandler) createIntegrationSecretHandler(w http.ResponseWriter, r *http.Request, payload dto.UpsertIntegrationSecretDto) {
	integrationIdStr := r.PathValue("integrationId")
	integrationId, err := uuid.Parse(integrationIdStr)
	if err != nil {
		res := api.ApiResponse{
			Code:    400,
			Errors:  err,
			Message: "Malformed integration id",
		}
		api.ProcessUsecaseResponse(&res, w)
		return
	}

	body := &payload
	body.IntegrationId = integrationId

	result := h.usecase.UpsertIntegrationSecretUsecase(r.Context(), body)
	api.ProcessUsecaseResponse(result, w)
}

// Delete an integration secret
func (h *integrationsHandler) deleteIntegrationSecretHandler(w http.ResponseWriter, r *http.Request, payload dto.IntegrationSecretDto) {
	integrationIdStr := r.PathValue("integrationId")
	integrationId, err := uuid.Parse(integrationIdStr)
	if err != nil {
		res := api.ApiResponse{
			Code:    400,
			Errors:  err,
			Message: "Malformed integration id",
		}
		api.ProcessUsecaseResponse(&res, w)
		return
	}

	body := &payload
	body.IntegrationId = integrationId

	result := h.usecase.DeleteIntegrationSecretUsecase(r.Context(), body)
	api.ProcessUsecaseResponse(result, w)
}
