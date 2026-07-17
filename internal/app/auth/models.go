package auth

import (
	"time"

	"github.com/google/uuid"
)

type CredentialsUserRequest struct {
	Email         string `json:"email"`
	Password_hash string `json:"password_hash"`
}

type UpdateUserRequest struct {
	Id            uuid.UUID `json:"id"`
	Email         *string   `json:"email,omitempty"`
	Password_hash *string   `json:"password_hash,omitempty"`
}

type UserCreatedResponse struct {
	Id         uuid.UUID `json:"id"`
	Email      string    `json:"email"`
	Created_at time.Time `json:"created_at"`
	Updated_at time.Time `json:"updated_at"`
}

type UserDataResponse struct {
	Id            uuid.UUID `json:"id"`
	Email         string    `json:"email"`
	Password_hash string    `json:"password_hash"`
	Created_at    time.Time `json:"created_at"`
	Updated_at    time.Time `json:"updated_at"`
}
