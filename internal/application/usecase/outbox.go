package usecase

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stivo-m/api.kodiflow.com/internal/infrastructure/database"
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
	case "waitlist.created":
		return p.handleWaitlistCreated(ctx, q, evt)

	default:
		log.Printf("unknown event type: %s", evt.EventType)
		return nil
	}
}

// Waitlist created handler
func (p *OutboxProcessor) handleWaitlistCreated(
	ctx context.Context,
	q *database.Queries,
	evt database.OutboxEvent,
) error {

	log.Printf("event %s has started processing ....", evt.ID)
	var payload struct {
		Email  string `json:"email"`
		Source string `json:"source"`
	}

	if err := json.Unmarshal(evt.Payload, &payload); err != nil {
		return err
	}

	return nil
}

