package events

const WaitlistAggregateType string = "waitlists"

type WaitlistEvent string

var (
	WaitlistCreatedEvent    WaitlistEvent = "waitlist.created"
	WaitlistInviteEvent     WaitlistEvent = "waitlist.invited"
	WaitlistRegisteredEvent WaitlistEvent = "waitlist.registered"
)

type WaitlistEventPayload struct {
	Email string `json:"email"`
}
