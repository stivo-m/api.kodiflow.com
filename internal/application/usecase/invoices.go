package usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"math"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stivo-m/api.kodiflow.com/internal/application/dto"
	"github.com/stivo-m/api.kodiflow.com/internal/domain/events"
	"github.com/stivo-m/api.kodiflow.com/internal/infrastructure/database"
	"github.com/stivo-m/api.kodiflow.com/pkg/api"
	"github.com/stivo-m/api.kodiflow.com/pkg/helpers"
)

type InvoiceUsecase struct {
	queries *database.Queries
	pool    *pgxpool.Pool
	logger  *slog.Logger
}

// New invoice usecase
func NewInvoiceUsecase(queries *database.Queries, pool *pgxpool.Pool) *InvoiceUsecase {
	logger := slog.Default().WithGroup("metadata").With(slog.String("type", "usecase"), slog.String("name", "invoices"))
	return &InvoiceUsecase{queries: queries, pool: pool, logger: logger}
}

// Creates a new invoice for the business
func (u *InvoiceUsecase) CreateInvoiceUsecase(ctx context.Context, payload *dto.CreateInvoiceDto) *api.ApiResponse {
	businessId, err := helpers.BusinessFromContext(ctx)
	if err != nil {
		u.logger.Error("failed to get business from context when creating an invoice", "error", err)
		return &api.ApiResponse{
			Code:    500,
			Errors:  err,
			Message: "Failed to create invoice",
		}
	}

	// Compute the total VAT, subtotal and total
	var itemParams []database.AddLineInvoiceLineItemParams
	total, subtotal, vatTotal := 0.0, 0.0, 0.0
	for _, item := range payload.LineItems {

		itemTotal := math.Round(float64(item.Quantity)*item.UnitPrice*100) / 100
		itemVatTotal := math.Round((item.VatRate/100)*itemTotal*100) / 100

		lineTotal := itemTotal + itemVatTotal

		subtotal += itemTotal
		vatTotal += itemVatTotal
		total += lineTotal

		record := database.AddLineInvoiceLineItemParams{
			Description: pgtype.Text{Valid: true, String: item.Description},
			Quantity:    helpers.Float64ToNumeric(float64(item.Quantity)),
			UnitPrice:   helpers.Float64ToNumeric(item.UnitPrice),
			VatRate:     helpers.Float64ToNumeric(item.VatRate),
			VatAmount:   helpers.Float64ToNumeric(itemVatTotal),
			LineTotal:   helpers.Float64ToNumeric(lineTotal),
		}
		itemParams = append(itemParams, record)
	}

	// Use a db tx to add the invoice
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		u.logger.Error("failed to start transaction when creating an invoice", "error", err)
		return &api.ApiResponse{
			Code:    500,
			Errors:  err,
			Message: "Failed to create invoice",
		}
	}

	defer tx.Rollback(ctx)
	qtx := u.queries.WithTx(tx)

	kraPin, err := qtx.GetBusinessKraPin(ctx, businessId)
	if err != nil {
		u.logger.Error("failed to get business kra pin when creating invoice", "error", err)
		return &api.ApiResponse{
			Code:    500,
			Errors:  err,
			Message: "Failed to create invoice",
		}
	}

	invoiceParams := database.CreateInvoiceParams{
		BusinessID:    businessId,
		SellerPin:     kraPin,
		InvoiceNumber: payload.InvoiceNumber,
		Subtotal:      helpers.Float64ToNumeric(subtotal),
		VatTotal:      helpers.Float64ToNumeric(vatTotal),
		Total:         helpers.Float64ToNumeric(total),
	}
	invoice, err := qtx.CreateInvoice(ctx, invoiceParams)
	if err != nil {
		u.logger.Error("failed to create an invoice record", "error", err)
		return &api.ApiResponse{
			Code:    500,
			Errors:  err,
			Message: "Failed to create invoice",
		}
	}

	// Use the db tx to add the line items
	for i := range itemParams {
		itemParams[i].InvoiceID = invoice.ID
	}
	_, err = qtx.AddLineInvoiceLineItem(ctx, itemParams)
	if err != nil {
		u.logger.Error("failed to insert line items for invoice", "invoice_id", invoice.ID, "error", err)
		return &api.ApiResponse{
			Code:    500,
			Errors:  err,
			Message: "Failed to create invoice",
		}
	}

	eventPayload := events.InvoiceEventPayload{
		ID:         invoice.ID,
		BusinessId: invoice.BusinessID,
	}

	jsonPayload, _ := json.Marshal(eventPayload)
	eventParams := database.CreateOutboxEventParams{
		AggregateType: string(events.InvoiceAggregateType),
		AggregateID:   invoice.ID,
		EventType:     string(events.InvoiceCreatedEvent),
		Payload:       jsonPayload,
	}

	err = qtx.CreateOutboxEvent(ctx, eventParams)
	if err != nil {
		u.logger.Error("failed to insert created event for invoice", "invoice_id", invoice.ID, "error", err)
		return &api.ApiResponse{
			Code:    500,
			Errors:  err,
			Message: "Failed to create invoice",
		}
	}

	tx.Commit(ctx)

	record := dto.InvoiceRecord{
		ID:            invoice.ID.String(),
		InvoiceNumber: invoice.InvoiceNumber,
		Currency:      invoice.Currency.String,
		CreatedAt:     &invoice.CreatedAt.Time,
		UpdatedAt:     &invoice.UpdatedAt.Time,
		IssuedAt:      &invoice.IssuedAt.Time,
		Status:        invoice.Status,
		Subtotal:      helpers.NumericToFloat64(invoice.Subtotal),
		VatTotal:      helpers.NumericToFloat64(invoice.VatTotal),
		Total:         helpers.NumericToFloat64(invoice.Total),
	}

	return &api.ApiResponse{
		Code:    201,
		Data:    record,
		Message: "Invoice created successfully",
	}

}

// Gets invoices for the business
func (u *InvoiceUsecase) ListInvociesUsecase(ctx context.Context) *api.ApiResponse {
	businessId, err := helpers.BusinessFromContext(ctx)
	if err != nil {
		u.logger.Error("failed to get business from context when creating an invoice", "error", err)
		return &api.ApiResponse{
			Code:    403,
			Errors:  err,
			Message: "Forbidden",
		}
	}

	rows, err := u.queries.ListBusinessInvoices(ctx, businessId)
	if err != nil && err != sql.ErrNoRows {
		u.logger.Error("failed to get invoices for business", "business_id", businessId, "error", err)
		return &api.ApiResponse{
			Code:    500,
			Errors:  err,
			Message: "Failed to get business invoices",
		}
	}

	var records []dto.InvoiceRecord
	for _, item := range rows {
		record := dto.InvoiceRecord{
			ID:            item.ID.String(),
			InvoiceNumber: item.InvoiceNumber,
			Currency:      item.Currency.String,
			CreatedAt:     &item.CreatedAt.Time,
			UpdatedAt:     &item.UpdatedAt.Time,
			IssuedAt:      &item.IssuedAt.Time,
			Status:        item.Status,
			Subtotal:      helpers.NumericToFloat64(item.Subtotal),
			VatTotal:      helpers.NumericToFloat64(item.VatTotal),
			Total:         helpers.NumericToFloat64(item.Total),
		}
		records = append(records, record)
	}

	return &api.ApiResponse{
		Code:    200,
		Data:    records,
		Message: "Invoice records fetched successfully",
	}
}
