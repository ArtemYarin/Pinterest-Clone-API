package rabbitmq

import (
	"time"

	"github.com/google/uuid"
)

// UserRegisteredEvent is published by auth-service after a new user signs up,
// and consumed by profile-service to create the corresponding profile row.
type UserRegisteredEvent struct {
	UserID            uuid.UUID `json:"user_id"`
	Email             string    `json:"email"`
	SuggestedUsername string    `json:"suggested_username"`
	OccurredAt        time.Time `json:"occurred_at"`
}
