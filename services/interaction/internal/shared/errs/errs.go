package errs

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Sentinel errors

var ErrForbidden = errors.New("forbidden error")
var ErrBadRequest = errors.New("bad request error")
var ErrUnauthorized = errors.New("unauthorized error")
var ErrLikeNotFound = errors.New("like not found error")
var ErrInternalServer = errors.New("internal server error")

// var ErrEmailExists = errors.New("conflict error")
// var ErrCommentNotFound = errors.New("comment not found error")

// Validation error
type ErrValidation struct {
	Err     error             `json:"-"`
	Details map[string]string `json:"details"`
}

func (e *ErrValidation) Error() string {
	return fmt.Sprintf("error: %s", e.Err)
}

func (e *ErrValidation) Unwrap() error {
	return e.Err
}

func NewValidationErr(dets map[string]string) *ErrValidation {
	return &ErrValidation{
		Err:     fmt.Errorf("validation error: %v", dets),
		Details: dets,
	}
}

// Helpers
func IsDuplicateErr(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return true
	}
	return false
}
func IsNotFoundErr(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
