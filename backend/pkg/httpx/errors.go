package httpx

import (
	"fmt"
)

// Error is an error that knows which HTTP status and message it maps to.
// Services declare their own sentinels with NewError and WriteJSONError
// renders them without knowing about the service.
type Error struct {
	Status  int
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

func NewError(status int, msg string) *Error {
	return &Error{Status: status, Message: msg}
}

// Common sentinel errors
var ErrBadRequest = NewError(400, "Bad request")
var ErrUnauthorized = NewError(401, "Unauthorized")
var ErrForbidden = NewError(403, "Forbidden")
var ErrInternalServer = NewError(500, "Internal server error")

// Validation error
type ValidationError struct {
	Err     error             `json:"-"`
	Details map[string]string `json:"details"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("error: %s", e.Err)
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

func NewValidationErr(dets map[string]string) *ValidationError {
	return &ValidationError{
		Err:     fmt.Errorf("validation error: %v", dets),
		Details: dets,
	}
}
