package pin

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Sentinel errors
var errForbidden = errors.New("forbidden error")
var errImageURLExists = errors.New("conflict error")
var errBadRequest = errors.New("bad request error")
var errUnauthorized = errors.New("unauthorized error")
var errPinNotFound = errors.New("pin not found error")
var errInternalServer = errors.New("internal server error")

// Validation error
type errValidation struct {
	Err     error             `json:"-"`
	Details map[string]string `json:"details"`
}

func (e *errValidation) Error() string {
	return fmt.Sprintf("error: %s", e.Err)
}

func (e *errValidation) Unwrap() error {
	return e.Err
}

func newValidationErr(dets map[string]string) *errValidation {
	return &errValidation{
		Err:     fmt.Errorf("validation error: %v", dets),
		Details: dets,
	}
}

// Helpers
func isDuplicateErr(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return true
	}
	return false
}
func isNotFoundErr(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
