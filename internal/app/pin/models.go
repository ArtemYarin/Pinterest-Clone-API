package pin

import (
	"time"

	"github.com/google/uuid"
)

type CreatePinRequest struct {
	Title       string  `json:"title"                 validate:"required,min=1,max=255"`
	Image_url   string  `json:"image_url"              validate:"required,url,max=255"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=1000"`
}

type UpdatePinRequest struct {
	Id          uuid.UUID `json:"-"`
	Title       *string   `json:"title,omitempty"       validate:"omitempty,min=1,max=255"`
	Image_url   *string   `json:"image_url,omitempty"   validate:"omitempty,url,max=255"`
	Description *string   `json:"description,omitempty" validate:"omitempty,max=1000"`
}

type PinResponse struct {
	Id          uuid.UUID `json:"id"`
	User_id     uuid.UUID `json:"user_id"`
	Title       string    `json:"title"`
	Image_url   string    `json:"image_url"`
	Description *string   `json:"description,omitempty"`
	Created_at  time.Time `json:"created_at"`
	Updated_at  time.Time `json:"updated_at"`
	Likes_count int       `json:"likes_count"`
}
