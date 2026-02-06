package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stivo-m/api.kodiflow.com/internal/application/dto"
	"github.com/stivo-m/api.kodiflow.com/internal/infrastructure/database"
	"github.com/stivo-m/api.kodiflow.com/pkg/api"
	"github.com/stivo-m/api.kodiflow.com/pkg/encryption"
	"github.com/stivo-m/api.kodiflow.com/pkg/helpers"
)

type BusinessIntegrationUsecase struct {
	queries *database.Queries
	pool    *pgxpool.Pool
}

// Creating a new business integration usecase
func NewBusinessIntegrationUsecase(
	queries *database.Queries,
	pool *pgxpool.Pool,
) *BusinessIntegrationUsecase {
	return &BusinessIntegrationUsecase{queries: queries, pool: pool}
}

// Creates a new business integration
func (u *BusinessIntegrationUsecase) CreateIntegrationUsecase(ctx context.Context, payload *dto.CreateIntegrationDto) *api.ApiResponse {
	businessId, err := helpers.BusinessFromContext(ctx)
	if err != nil {
		return &api.ApiResponse{
			Code:    403,
			Errors:  err,
			Message: "Unauthorized",
		}
	}

	params := database.CreateBusinessIntegrationParams{
		BusinessID:  businessId,
		Provider:    payload.Provider,
		Environment: payload.Environment,
	}

	record, err := u.queries.CreateBusinessIntegration(ctx, params)
	if err != nil {
		return &api.ApiResponse{
			Code:    500,
			Errors:  err,
			Message: "Failed to create business integration",
		}
	}

	out := dto.IntegrationRecord{
		ID:          record.ID.String(),
		Provider:    record.Provider,
		Environment: record.Environment,
		CreatedAt:   &record.CreatedAt.Time,
	}

	return &api.ApiResponse{
		Code:    201,
		Data:    out,
		Message: "Business integration created successfully",
	}
}

// Obtains integrations for a business
func (u *BusinessIntegrationUsecase) ListIntegrationsUsecase(ctx context.Context) *api.ApiResponse {
	businessId, err := helpers.BusinessFromContext(ctx)
	if err != nil {
		return &api.ApiResponse{
			Code:    403,
			Errors:  err,
			Message: "Unauthorized",
		}
	}

	rows, err := u.queries.ListBusinessIntegrations(ctx, businessId)
	if err != nil {
		return &api.ApiResponse{
			Code:    500,
			Errors:  err,
			Message: "Failed to list business integrations",
		}
	}

	var records []dto.IntegrationRecord
	for _, record := range rows {
		out := dto.IntegrationRecord{
			ID:          record.ID.String(),
			Provider:    record.Provider,
			Environment: record.Environment,
			CreatedAt:   &record.CreatedAt.Time,
		}
		records = append(records, out)
	}

	return &api.ApiResponse{
		Code:    200,
		Data:    records,
		Message: "Business integration fetched successfully",
	}
}

// Deletes a business integration
func (u *BusinessIntegrationUsecase) DeleteBusinessIntegrationUsecase(ctx context.Context, integrationId uuid.UUID) *api.ApiResponse {
	businessId, err := helpers.BusinessFromContext(ctx)
	if err != nil {
		return &api.ApiResponse{
			Code:    403,
			Errors:  err,
			Message: "Unauthorized",
		}
	}

	params := database.DeleteBusinessIntegrationParams{
		ID:         integrationId,
		BusinessID: businessId,
	}

	err = u.queries.DeleteBusinessIntegration(ctx, params)
	if err != nil {
		return &api.ApiResponse{
			Code:    500,
			Errors:  err,
			Message: "Failed to delete business integration",
		}
	}

	return &api.ApiResponse{
		Code:    200,
		Message: "Business integration deleted successfully",
	}
}

// Upsert an integration secret
func (u *BusinessIntegrationUsecase) UpsertIntegrationSecretUsecase(ctx context.Context, payload *dto.UpsertIntegrationSecretDto) *api.ApiResponse {
	businessId, err := helpers.BusinessFromContext(ctx)
	if err != nil {
		return &api.ApiResponse{
			Code:    403,
			Errors:  err,
			Message: "Unauthorized",
		}
	}

	encryptedValue, err := encryption.Encrypt(payload.KeyValue, businessId.String())
	if err != nil {
		return &api.ApiResponse{
			Code:    500,
			Errors:  err,
			Message: "Failed to encrypt secret",
		}
	}

	params := database.UpsertIntegrationSecretParams{
		EncryptedValue: encryptedValue,
		KeyName:        payload.KeyName,
		IntegrationID:  payload.IntegrationId,
	}

	_, err = u.queries.UpsertIntegrationSecret(ctx, params)
	if err != nil {
		return &api.ApiResponse{
			Code:    500,
			Errors:  err,
			Message: "Failed to upsert integration secret",
		}
	}

	return &api.ApiResponse{
		Code:    201,
		Message: "Integration secred securely saved",
	}
}

// Delete an integration secret
func (u *BusinessIntegrationUsecase) DeleteIntegrationSecretUsecase(ctx context.Context, payload *dto.IntegrationSecretDto) *api.ApiResponse {
	deleteParams := database.DeleteIntegrationSecretParams{
		KeyName:       payload.KeyName,
		IntegrationID: payload.IntegrationId,
	}

	err := u.queries.DeleteIntegrationSecret(ctx, deleteParams)
	if err != nil {
		return &api.ApiResponse{
			Code:    500,
			Errors:  err,
			Message: "Failed to get the integration secret",
		}
	}

	return &api.ApiResponse{
		Code:    200,
		Message: "Integration secret deleted successfully",
	}
}
