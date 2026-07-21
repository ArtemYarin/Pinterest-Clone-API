package auth

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var errBadRequest = errors.New("bad request error")
var errUserNotFound = errors.New("user not found error")
var errValidation = errors.New("validation error")
var errInternalServer = errors.New("internal server error")
var errUnauthorized = errors.New("unauthorized error")
var errForbidden = errors.New("forbidden error")
var errEmailExists = errors.New("conflict error")

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
