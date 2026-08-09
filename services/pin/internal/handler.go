package pin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/ArtemYarin/pinterest-clone-api/pkg/middleware"
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

func (h *PinHandler) GetPinByID(w http.ResponseWriter, r *http.Request) {
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

func (h *PinHandler) GetPins(w http.ResponseWriter, r *http.Request) {
	filters := ParseFilters(r)

	if err := filters.Validate(); err != nil {
		WriteJSONError(fmt.Errorf("validate query parameters: %w", err), w)
		return
	}

	pins, count, err := h.service.GetPins(r.Context(), filters)
	if err != nil {
		WriteJSONError(err, w)
	}

	// Response with metadata
	response := map[string]interface{}{
		"data": pins,
		"metadata": map[string]interface{}{
			"total":      count,
			"user_id":    filters.UserID,
			"search":     filters.Search,
			"sort_by":    filters.SortBy,
			"sort_order": filters.SortOrder,
			"limit":      filters.Limit,
			"offset":     filters.Offset,
		},
	}

	WriteJSON(w, http.StatusOK, response)
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

// Helpers
type PinFilters struct {
	UserID    string `json:"user_id,omitempty"`
	Search    string `json:"search,omitempty"`
	SortBy    string `json:"sort_by,omitempty"`
	SortOrder string `json:"sort_order,omitempty"` // asc, desc
	Limit     int    `json:"limit,omitempty"`      // max = 100, default = 20
	Offset    int    `json:"offset,omitempty"`
}

func (f *PinFilters) Validate() error {
	f.setDefault()

	// Validate SortBy - Whitelist
	allowedSortFields := map[string]bool{
		"title":      true,
		"created_at": true,
		"updated_at": true,
		"likes":      true,
		"":           true, // empty is allowed (will use default)
	}
	if !allowedSortFields[f.SortBy] {
		valErr := errValidation{
			Details: map[string]string{
				"invalid sort_by field": "Allowed: 'title', 'created_at', 'updated_at', 'likes' or empty",
			},
		}
		return fmt.Errorf("invalid sort_by field: %s. Allowed: 'title', 'created_at', 'updated_at', 'likes' or empty: %w", f.SortBy, &valErr)
	}

	// Validate SortOrder
	if f.SortOrder != "" && f.SortOrder != "asc" && f.SortOrder != "desc" {
		valErr := errValidation{
			Details: map[string]string{
				"invalid sort_order": "Must be 'asc', 'desc' or empty",
			},
		}
		return fmt.Errorf("invalid sort_order: %s. Must be 'asc', 'desc' or empty: %w", f.SortOrder, &valErr)
	}

	// Validate Limit
	if f.Limit < 1 {
		f.Limit = 20
	} else if f.Limit > 100 {
		f.Limit = 100
	}

	// Validate Offset (must be non-negative)
	if f.Offset < 0 {
		f.Offset = 0
	}

	return nil
}

func (f *PinFilters) setDefault() {
	if f.SortBy == "" {
		f.SortBy = "created_at"
	}

	if f.SortOrder == "" {
		f.SortOrder = "desc"
	}
}

func ParseFilters(r *http.Request) PinFilters {
	filters := PinFilters{
		UserID:    r.URL.Query().Get("user_id"),
		Search:    r.URL.Query().Get("search"),
		SortBy:    r.URL.Query().Get("sort_by"),
		SortOrder: r.URL.Query().Get("sort_order"),
	}

	if limit := r.URL.Query().Get("limit"); limit != "" {
		filters.Limit, _ = strconv.Atoi(limit)
	} else {
		filters.Limit = 20
	}

	if offset := r.URL.Query().Get("offset"); offset != "" {
		filters.Offset, _ = strconv.Atoi(offset)
	}

	return filters
}
