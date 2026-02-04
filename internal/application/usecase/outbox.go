package usecase

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stivo-m/api.kodiflow.com/internal/domain/events"
	"github.com/stivo-m/api.kodiflow.com/internal/infrastructure/database"
	"github.com/stivo-m/api.kodiflow.com/pkg/hashing"
)

type OutboxProcessor struct {
	Queries       *database.Queries
	Pool          *pgxpool.Pool
	PollInterval  time.Duration
	BatchSize     int
	MaxRetryCount int
}

// Starts the actual processing for the worker service
func (p *OutboxProcessor) Start(ctx context.Context) {
	ticker := time.NewTicker(p.PollInterval)
	defer ticker.Stop()

	log.Println("outbox worker started")

	for {
		select {
		case <-ctx.Done():
			log.Println("outbox worker shutting down")
			return

		case <-ticker.C:
			if err := p.processOnce(ctx); err != nil {
				log.Printf("outbox processing error: %v\n", err)
			}
		}
	}
}

//

func (p *OutboxProcessor) processOnce(ctx context.Context) error {
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)
	q := p.Queries.WithTx(tx)

	events, err := q.FetchPendingOutboxEvents(ctx, int32(p.BatchSize))
	if err != nil {
		return err
	}

	for _, evt := range events {
		if err := q.MarkOutboxEventProcessing(ctx, evt.ID); err != nil {
			return err
		}

		if err := p.handleEvent(ctx, q, evt); err != nil {
			log.Printf("event %s failed: %v", evt.ID, err)

			if evt.RetryCount >= int32(p.MaxRetryCount) {
				lastErr := err.Error()

				_ = q.MarkOutboxEventDead(ctx, database.MarkOutboxEventDeadParams{
					ID:        evt.ID,
					LastError: pgtype.Text{String: lastErr, Valid: true},
				})
			} else {
				lastErr := err.Error()
				_ = q.MarkOutboxEventFailed(ctx, database.MarkOutboxEventFailedParams{
					ID:        evt.ID,
					LastError: pgtype.Text{String: lastErr, Valid: true},
				})
			}
			continue
		}

		if err := q.MarkOutboxEventProcessed(ctx, evt.ID); err != nil {
			return err
		}

		log.Printf("event %s has been processed successfully", evt.ID)
	}

	return tx.Commit(ctx)
}

// Handles events based on their types

func (p *OutboxProcessor) handleEvent(
	ctx context.Context,
	q *database.Queries,
	evt database.OutboxEvent,
) error {
	switch evt.EventType {
	case string(events.UserCreatedEvent):
		return p.handleUserCreated(ctx, q, evt)
	case string(events.UserForgotPasswordEvent):
		return p.handleUserForgotPassword(ctx, q, evt)

	default:
		log.Printf("unknown event type: %s", evt.EventType)
		return nil
	}
}

// Handle new user accounts
func (p *OutboxProcessor) handleUserCreated(
	ctx context.Context,
	q *database.Queries,
	evt database.OutboxEvent,
) error {

	var payload events.UserEventPayload
	if err := json.Unmarshal(evt.Payload, &payload); err != nil {
		return err
	}

	code, err := hashing.GenerateOTP(6)
	if err != nil {
		return err
	}

	params := database.StoreVerificationRecordParams{
		UserID:           evt.AggregateID,
		VerificationType: "email_verification",
		VerificationCode: code,
	}

	// TODO: Send verification email

	return q.StoreVerificationRecord(ctx, params)
}

// Handle user forgot password
func (p *OutboxProcessor) handleUserForgotPassword(
	ctx context.Context,
	q *database.Queries,
	evt database.OutboxEvent,
) error {

	var payload events.UserEventPayload
	if err := json.Unmarshal(evt.Payload, &payload); err != nil {
		return err
	}

	code, err := hashing.GenerateOTP(6)
	if err != nil {
		return err
	}

	params := database.StoreVerificationRecordParams{
		UserID:           evt.AggregateID,
		VerificationType: "reset_password",
		VerificationCode: code,
	}

	// TODO: Send verification email

	return q.StoreVerificationRecord(ctx, params)
}
