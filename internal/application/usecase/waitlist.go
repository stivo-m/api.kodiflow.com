package usecase

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stivo-m/api.kodiflow.com/internal/application/dto"
	"github.com/stivo-m/api.kodiflow.com/internal/domain/events"
	"github.com/stivo-m/api.kodiflow.com/internal/infrastructure/database"
	"github.com/stivo-m/api.kodiflow.com/pkg/api"
)

type WaitlistUsecase struct {
	queries *database.Queries
	pool    *pgxpool.Pool
}

// Instantiates a new waitlist usecase
func NewWaitlistUsecase(queries *database.Queries, pool *pgxpool.Pool) *WaitlistUsecase {
	return &WaitlistUsecase{queries: queries, pool: pool}
}

// Handles the creation of a new waitlist record and sending them a notification
func (u *WaitlistUsecase) AddUserToWaitlist(ctx context.Context, payload *dto.CreateWaitlistDto) *api.ApiResponse {

	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return &api.ApiResponse{
			Code:    500,
			Message: "Unable to start transaction to create the testimonial",
			Errors:  err,
		}
	}

	defer tx.Rollback(ctx)
	qtx := u.queries.WithTx(tx)

	addToWaitlistParam := database.AddToWaitlistParams{
		Source: pgtype.Text{Valid: true, String: payload.Source},
		Email:  payload.Email,
	}

	waitlist, err := qtx.AddToWaitlist(ctx, addToWaitlistParam)
	if err != nil {
		return &api.ApiResponse{
			Code:    500,
			Errors:  err,
			Message: "Failed to add user to waitlist",
		}
	}

	eventPayload := events.WaitlistEventPayload{
		Email: waitlist.Email,
	}

	jsonPayload, err := json.Marshal(eventPayload)
	if err != nil {
		return &api.ApiResponse{
			Code:    500,
			Errors:  err,
			Message: "Failed to marchal json event payload",
		}
	}

	outboxParams := database.CreateOutboxEventParams{
		AggregateType: events.WaitlistAggregateType,
		AggregateID:   waitlist.ID,
		EventType:     string(events.WaitlistCreatedEvent),
		Payload:       jsonPayload,
	}

	err = qtx.CreateOutboxEvent(ctx, outboxParams)
	if err != nil {
		return &api.ApiResponse{
			Code:    500,
			Errors:  err,
			Message: "Failed to create outbox event for new waitlist",
		}
	}

	tx.Commit(ctx)

	return &api.ApiResponse{
		Code:    201,
		Message: "User has been added to the waitlists",
	}
}
