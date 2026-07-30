package auth

import (
	"time"

	"github.com/google/uuid"
)

type CredentialsUserRequest struct {
	Email         string `json:"email"         validate:"required,email"`
	Password_hash string `json:"password_hash" validate:"required,min=8,max=50"`
}

type UpdateUserRequest struct {
	Id            uuid.UUID `json:"id"`
	Email         *string   `json:"email,omitempty" validate:"omitempty,email"`
	Password_hash *string   `json:"password_hash,omitempty" validate:"omitempty,min=8,max=50"`
}

type UserResponse struct {
	Id         uuid.UUID `json:"id"`
	Email      string    `json:"email"`
	Created_at time.Time `json:"created_at"`
	Updated_at time.Time `json:"updated_at"`
}

type UserWithPasswordResponse struct {
	Id            uuid.UUID `json:"id"`
	Email         string    `json:"email"`
	Password_hash string    `json:"password_hash"`
	Created_at    time.Time `json:"created_at"`
	Updated_at    time.Time `json:"updated_at"`
}

type UserWithTokenResponse struct {
	UserResponse `json:"user"`
	Token        string `json:"token"`
}
