package pin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ArtemYarin/pinterest-clone-api/services/pin-service/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PinHandler struct {
	service PinService
}

func NewPinHandler(s PinService) PinHandler {
	return PinHandler{service: s}
}

func (h *PinHandler) CreatePin(w http.ResponseWriter, r *http.Request) {
	// Decode body
	var pin CreatePinRequest
	if err := DecodeJSON(&pin, r); err != nil {
		WriteJSONError(err, w)
		return
	}

	// Security
	claims, ok := middleware.GetUserClaims(r)
	if !ok {
		WriteJSONError(errUnauthorized, w)
		return
	}

	// Service call
	resp, err := h.service.CreatePin(r.Context(), claims.UserID, pin)
	if err != nil {
		WriteJSONError(fmt.Errorf("create pin: %w", err), w)
		return
	}

	// Response writing
	WriteJSON(w, http.StatusCreated, resp)
}

func (h *PinHandler) GetPin(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if !IsValidUUID(id) {
		WriteJSONError(fmt.Errorf("get pin: %w", errBadRequest), w)
		return
	}

	// Service call
	resp, err := h.service.GetPinByID(r.Context(), id)
	if err != nil {
		WriteJSONError(fmt.Errorf("get pin: %w", err), w)
		return
	}

	// Response writing
	WriteJSON(w, 200, resp)
}

func (h *PinHandler) UpdatePin(w http.ResponseWriter, r *http.Request) {
	// Get id from url parameters
	id := chi.URLParam(r, "id")

	if !IsValidUUID(id) {
		WriteJSONError(fmt.Errorf("update pin: %w", errBadRequest), w)
		return
	}

	// Decode body
	var pin UpdatePinRequest
	if err := DecodeJSON(&pin, r); err != nil {
		WriteJSONError(err, w)
		return
	}

	// Security
	claims, ok := middleware.GetUserClaims(r)
	if !ok {
		WriteJSONError(errUnauthorized, w)
		return
	}

	// Service call
	err := h.service.UpdatePin(r.Context(), id, claims.UserID, pin)
	if err != nil {
		WriteJSONError(fmt.Errorf("update pin: %w", err), w)
		return
	}

	// Response writing
	WriteJSON(w, 200, "Updated successfully")
}

func (h *PinHandler) DeletePin(w http.ResponseWriter, r *http.Request) {
	// Get id from url parameters
	id := chi.URLParam(r, "id")

	if !IsValidUUID(id) {
		WriteJSONError(fmt.Errorf("delete pin: %w", errBadRequest), w)
		return
	}

	// Security
	claims, ok := middleware.GetUserClaims(r)
	if !ok {
		WriteJSONError(errUnauthorized, w)
		return
	}

	// Service call
	err := h.service.DeletePin(r.Context(), id, claims.UserID)
	if err != nil {
		WriteJSONError(fmt.Errorf("delete pin: %w", err), w)
		return
	}

	// Response writing
	WriteJSON(w, 200, "Deleted successfully")
}

// Health
func Health(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := http.StatusOK
		status := "ok"
		postgresStatus := "healthy"
		if err := db.Ping(r.Context()); err != nil {
			code = http.StatusServiceUnavailable
			postgresStatus = "unhealthy"
			status = "unhealthy"
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		json.NewEncoder(w).Encode(map[string]any{
			"service":    "pis-service",
			"status":     status,
			"PostgreSQL": postgresStatus,
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
		})
	}
}
