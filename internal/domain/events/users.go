package events

import "github.com/google/uuid"

var (
	UserCreatedEvent         DomainEvent = "users.created"
	UserForgotPasswordEvent  DomainEvent = "users.forgot_password"
	UserResetPasswordEvent   DomainEvent = "users.reset_password"
	UserSignedInEvent        DomainEvent = "users.signed_in"
	UserAddedToBusinessEvent DomainEvent = "users.added_to_business"

	UserAggregateType AggregateType = "users"
)

type UserEventPayload struct {
	ID         uuid.UUID `json:"id"`
	Email      string    `json:"email"`
	Phone      string    `json:"phone"`
	BusinessId uuid.UUID `json:"business_id"`
}
