package likes

import (
	"time"

	"github.com/google/uuid"
)

type LikeResponse struct {
	UserID    uuid.UUID `json:"user_id"`
	PinID     uuid.UUID `json:"pin_id"`
	CreatedAt time.Time `json:"created_at"`
}
