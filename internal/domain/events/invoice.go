package events

import "github.com/google/uuid"

var (
	InvoiceCreatedEvent DomainEvent = "invoice.created"
	InvoiceUpdatedEvent DomainEvent = "invoice.created"

	InvoiceAggregateType AggregateType = "invoice"
)

type InvoiceEventPayload struct {
	ID         uuid.UUID `json:"id"`
	BusinessId uuid.UUID `json:"business_id"`
}
