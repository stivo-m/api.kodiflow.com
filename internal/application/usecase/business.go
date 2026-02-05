package usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stivo-m/api.kodiflow.com/internal/application/dto"
	"github.com/stivo-m/api.kodiflow.com/internal/domain/events"
	"github.com/stivo-m/api.kodiflow.com/internal/infrastructure/database"
	"github.com/stivo-m/api.kodiflow.com/pkg/api"
	"github.com/stivo-m/api.kodiflow.com/pkg/helpers"
)

type BusinessUsecase struct {
	queries *database.Queries
	pool    *pgxpool.Pool
}

// new business usecase
func NewBusinessUsecase(queries *database.Queries, pool *pgxpool.Pool) *BusinessUsecase {
	return &BusinessUsecase{queries: queries, pool: pool}
}

// Checks if a given user is part of a given business
func (u *BusinessUsecase) checkUserBusinessOwnership(ctx context.Context, businessId uuid.UUID) (bool, error) {
	userId, err := helpers.GetUserFromContext(ctx)
	if err != nil {
		return false, err
	}

	params := database.CheckIfUserIsPartOfBusinessParams{
		BusinessID: businessId,
		UserID:     userId,
	}
	ok, err := u.queries.CheckIfUserIsPartOfBusiness(ctx, params)
	if err != nil {
		return false, err
	}

	if !ok {
		return false, errors.New("user is not part of this businessess")
	}

	return true, nil
}

// Creates a new business on behalf of the authenticated user and adds them as the admin
func (u *BusinessUsecase) CreateBusinessUsecase(ctx context.Context, payload *dto.CreateBusinessDto) *api.ApiResponse {
	userId, err := helpers.GetUserFromContext(ctx)
	if err != nil {
		return &api.ApiResponse{
			Code:    401,
			Errors:  err,
			Message: "Unauthenticated",
		}
	}

	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return &api.ApiResponse{
			Code:    500,
			Errors:  err,
			Message: "Unable to handle tx for creating a business",
		}
	}

	defer tx.Rollback(ctx)
	qtx := u.queries.WithTx(tx)

	businessParams := database.CreateBusinessParams{
		LegalName:    payload.LegalName,
		BusinessType: pgtype.Text{String: payload.BusinessType, Valid: true},
		TradingName:  pgtype.Text{String: payload.TradingName, Valid: true},
		KraPin:       payload.KraPin,
		Email:        pgtype.Text{String: payload.Email, Valid: true},
		Phone:        pgtype.Text{String: payload.Phone, Valid: true},
		Country:      pgtype.Text{String: payload.Country, Valid: true},
	}
	business, err := qtx.CreateBusiness(ctx, businessParams)
	if err != nil {
		return &api.ApiResponse{
			Code:    500,
			Errors:  err,
			Message: "Failed to create business",
		}
	}

	addUserParam := database.AddUserToBusinessParams{
		BusinessID: business.ID,
		UserID:     userId,
		Role:       pgtype.Text{String: "admin", Valid: true},
	}
	err = qtx.AddUserToBusiness(ctx, addUserParam)
	if err != nil {
		return &api.ApiResponse{
			Code:    500,
			Errors:  err,
			Message: "Failed to add user as admin to business",
		}
	}

	user, err := qtx.FindUserById(ctx, userId)
	if err != nil {
		return &api.ApiResponse{
			Code:    500,
			Errors:  err,
			Message: "Failed to obtain user to send notification to",
		}
	}

	p := events.UserEventPayload{
		ID:         user.ID,
		Email:      user.Email,
		Phone:      user.Phone.String,
		BusinessId: business.ID,
	}

	jsonPayload, _ := json.Marshal(p)

	eventPayload := database.CreateOutboxEventParams{
		AggregateType: string(events.UserAggregateType),
		AggregateID:   user.ID,
		EventType:     string(events.UserAddedToBusinessEvent),
		Payload:       jsonPayload,
	}

	err = qtx.CreateOutboxEvent(ctx, eventPayload)
	if err != nil {
		return &api.ApiResponse{
			Code:    500,
			Errors:  err,
			Message: "Failed to create outbox event when adding user to a business",
		}
	}

	tx.Commit(ctx)

	return &api.ApiResponse{
		Code:    201,
		Message: "Business created successfully",
	}
}

// Adds a user to the business
func (u *BusinessUsecase) AddUserToBusinessUsecase(ctx context.Context, payload *dto.AddUserToBusinessDto) *api.ApiResponse {
	owns, err := u.checkUserBusinessOwnership(ctx, payload.BusinessId)
	if err != nil || !owns {
		return &api.ApiResponse{
			Code:    403,
			Errors:  err,
			Message: "Unable to verify if authenticated user owns the current business",
		}
	}

	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return &api.ApiResponse{
			Code:    500,
			Errors:  err,
			Message: "Unable to handle tx for adding a user to a business",
		}
	}

	defer tx.Rollback(ctx)
	qtx := u.queries.WithTx(tx)

	params := database.AddUserToBusinessParams{
		UserID:     payload.UserId,
		BusinessID: payload.BusinessId,
		Role:       pgtype.Text{String: payload.Role, Valid: true},
	}

	err = qtx.AddUserToBusiness(ctx, params)
	if err != nil {
		return &api.ApiResponse{
			Code:    500,
			Errors:  err,
			Message: "Failed to add user to business",
		}
	}

	user, err := qtx.FindUserById(ctx, params.UserID)
	if err != nil {
		return &api.ApiResponse{
			Code:    500,
			Errors:  err,
			Message: "Failed to obtain user to send notification to",
		}
	}

	p := events.UserEventPayload{
		ID:         user.ID,
		Email:      user.Email,
		Phone:      user.Phone.String,
		BusinessId: payload.BusinessId,
	}

	jsonPayload, _ := json.Marshal(p)

	eventPayload := database.CreateOutboxEventParams{
		AggregateType: string(events.UserAggregateType),
		AggregateID:   user.ID,
		EventType:     string(events.UserAddedToBusinessEvent),
		Payload:       jsonPayload,
	}

	err = qtx.CreateOutboxEvent(ctx, eventPayload)
	if err != nil {
		return &api.ApiResponse{
			Code:    500,
			Errors:  err,
			Message: "Failed to create outbox event when adding user to a business",
		}
	}

	tx.Commit(ctx)

	return &api.ApiResponse{
		Code:    201,
		Message: "User has been added to the business",
	}

}

// Lists businessess that a user is part of
func (u *BusinessUsecase) ListBusinessessForUserUsecase(ctx context.Context) *api.ApiResponse {

	userId, err := helpers.GetUserFromContext(ctx)
	if err != nil {
		return &api.ApiResponse{
			Code:    403,
			Message: "Forbidden",
		}
	}

	rows, err := u.queries.ListBusinessesForUser(ctx, userId)
	if err != nil && err != sql.ErrNoRows {
		return &api.ApiResponse{
			Code:    500,
			Errors:  err,
			Message: "Failed to load businessess for current user",
		}
	}

	var records []dto.BusinessRecord
	for _, item := range rows {
		record := dto.BusinessRecord{
			ID:           item.ID.String(),
			LegalName:    item.LegalName,
			KraPin:       item.KraPin,
			BusinessType: item.BusinessType.String,
			Email:        item.Email.String,
			Phone:        item.Phone.String,
			Country:      item.Country.String,
			CreatedAt:    &item.CreatedAt.Time,
			Status:       item.Status.(string),
		}

		records = append(records, record)
	}

	return &api.ApiResponse{
		Code: 200,
		Data: dto.BusinessesResponse{
			Records: records,
		},
		Message: "Business records fetched successfully",
	}
}

// Creates a business KYC
func (u *BusinessUsecase) CreateBusinessKycUsecase(ctx context.Context, businessId uuid.UUID, payload *dto.CreateBusinessKycDto) *api.ApiResponse {
	owns, err := u.checkUserBusinessOwnership(ctx, businessId)
	if err != nil || !owns {
		return &api.ApiResponse{
			Code:    403,
			Errors:  err,
			Message: "Unable to verify user is part of this business",
		}
	}

	params := database.CreateBusinessKYCParams{
		BusinessID:   businessId,
		DocumentType: pgtype.Text{String: payload.DocumentType, Valid: true},
		DocumentUrl:  pgtype.Text{String: payload.DocumentUrl, Valid: true},
	}

	item, err := u.queries.CreateBusinessKYC(ctx, params)
	if err != nil {
		return &api.ApiResponse{
			Code:    500,
			Errors:  err,
			Message: "Failed to create business kyc",
		}
	}

	record := dto.BusinessKycRecord{
		ID:           item.ID.String(),
		DocumentType: item.DocumentType.String,
		DocumentUrl:  item.DocumentUrl.String,
		Status:       item.StatusText.String,
		StatusText:   item.StatusText.String,
		VerifiedAt:   &item.VerifiedAt.Time,
		CreatedAt:    &item.CreatedAt.Time,
	}

	return &api.ApiResponse{
		Code:    201,
		Data:    record,
		Message: "Business kyc created successfully",
	}
}

// Lists kyc records for a business
func (u *BusinessUsecase) ListKycForBusinessUsecase(ctx context.Context, businessId uuid.UUID) *api.ApiResponse {
	owns, err := u.checkUserBusinessOwnership(ctx, businessId)
	if err != nil || !owns {
		return &api.ApiResponse{
			Code:    403,
			Errors:  err,
			Message: "Unable to verify user is part of this business",
		}
	}

	rows, err := u.queries.GetBusinessKYC(ctx, businessId)
	if err != nil && err != sql.ErrNoRows {
		return &api.ApiResponse{
			Code:    500,
			Errors:  err,
			Message: "Failed to load business kyc",
		}
	}

	var records []dto.BusinessKycRecord

	for _, item := range rows {
		record := dto.BusinessKycRecord{
			ID:           item.ID.String(),
			DocumentType: item.DocumentType.String,
			DocumentUrl:  item.DocumentUrl.String,
			Status:       item.StatusText.String,
			StatusText:   item.StatusText.String,
			VerifiedAt:   &item.VerifiedAt.Time,
			CreatedAt:    &item.CreatedAt.Time,
		}
		records = append(records, record)
	}

	return &api.ApiResponse{
		Code: 200,
		Data: dto.BusinessKycResponse{
			Records: records,
		},
		Message: "Fetched kyc records successfully",
	}
}
