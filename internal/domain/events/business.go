package events

import "github.com/google/uuid"

var (
	BusinessCreatedEvent DomainEvent = "businesses.created"
	BusinessUpdatedEvent DomainEvent = "businesses.updated"
	BusinessDeletedEvent DomainEvent = "businesses.deleted"
)

type BusinessEventPayload struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
	Phone string    `json:"phone"`
}
