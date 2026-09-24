package profile

import (
	"time"

	"github.com/google/uuid"
)

type Profile struct {
	UserID       uuid.UUID `json:"user_id"`
	Username     string    `json:"username"`
	Bio          string    `json:"bio"`
	AvatarKey    *string   `json:"-"`
	AvatarStatus string    `json:"avatar_status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UpdateProfileRequest struct {
	UserID   uuid.UUID `json:"-"`
	Username *string   `json:"username,omitempty" validate:"omitempty,min=3,max=30"`
	Bio      *string   `json:"bio,omitempty"      validate:"omitempty,max=500"`
}

type ProfileResponse struct {
	UserID    uuid.UUID `json:"user_id"`
	Username  string    `json:"username"`
	Bio       string    `json:"bio"`
	AvatarURL string    `json:"avatar_url,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UploadURLResponse struct {
	UploadURL string `json:"upload_url"`
}
