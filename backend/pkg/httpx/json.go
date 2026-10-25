package httpx

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func WriteJSON(w http.ResponseWriter, status int, v any) {
	// Encode before writing headers so a failure can still produce a clean error response
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(v); err != nil {
		WriteJSONError(fmt.Errorf("encode response: %w", err), w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(buf.Bytes())
}

func DecodeJSON(dst any, r *http.Request) error {
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		var syntaxErr *json.SyntaxError
		var typeErr *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxErr):
			return fmt.Errorf("malformed json: %w", ErrBadRequest)
		case errors.As(err, &typeErr):
			return fmt.Errorf("invalid field type: %w", ErrBadRequest)
		default:
			return fmt.Errorf("invalid request body: %w", ErrBadRequest)
		}
	}
	return nil
}

func WriteJSONError(err error, w http.ResponseWriter) {
	log.Printf("Error: %v", err.Error())

	var valErr *ValidationError
	var appErr *Error
	switch {
	case errors.As(err, &valErr):
		writeBody(w, http.StatusUnprocessableEntity, map[string]any{
			"code":    "422",
			"Details": valErr.Details,
		})
	case errors.As(err, &appErr):
		writeBody(w, appErr.Status, map[string]string{
			"code":    strconv.Itoa(appErr.Status),
			"message": appErr.Message,
		})
	default:
		writeBody(w, http.StatusInternalServerError, map[string]string{
			"code":    "500",
			"message": ErrInternalServer.Message,
		})
	}
}

// Helpers
func writeBody(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
