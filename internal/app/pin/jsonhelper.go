package pin

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		WriteJSONError(err, w)
	}
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
			return fmt.Errorf("malformed json: %w", errBadRequest)
		case errors.As(err, &typeErr):
			return fmt.Errorf("invalid field type: %w", errBadRequest)
		default:
			return fmt.Errorf("invalid request body: %w", errBadRequest)
		}
	}
	return nil
}

func WriteJSONError(err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")

	log.Printf("Error: %v", err.Error())

	var valErr *errValidation
	switch {
	case errors.As(err, &valErr):
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]any{
			"code":    "422",
			"Details": valErr.Details,
		})
		return
	case errors.Is(err, errImageURLExists):
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{
			"code":    "409",
			"message": "Image URL already exists",
		})
		return
	case errors.Is(err, errBadRequest):
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"code":    "400",
			"message": "Bad request",
		})
		return
	case errors.Is(err, errPinNotFound):
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"code":    "404",
			"message": "pin not found",
		})
		return
	case errors.Is(err, errInternalServer):
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"code":    "500",
			"message": "Internal server error",
		})
		return
	case errors.Is(err, errUnauthorized):
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"code":    "401",
			"message": "Unauthorized",
		})
		return
	case errors.Is(err, errForbidden):
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{
			"code":    "403",
			"message": "Forbidden",
		})
		return
	default:
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"code":    "500",
			"message": "Internal server error",
		})
	}
}
