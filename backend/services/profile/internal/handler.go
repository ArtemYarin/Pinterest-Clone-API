package profile

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ArtemYarin/pinterest-clone-api/pkg/httpx"
	"github.com/ArtemYarin/pinterest-clone-api/pkg/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProfileHandler struct {
	service ProfileService
}

func NewProfileHandler(s ProfileService) ProfileHandler {
	return ProfileHandler{service: s}
}

func (h *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")

	resp, err := h.service.GetProfile(r.Context(), userID)
	if err != nil {
		httpx.WriteJSONError(fmt.Errorf("get profile: %w", err), w)
		return
	}

	httpx.WriteJSON(w, 200, resp)
}

func (h *ProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")

	if !httpx.IsValidUUID(userID) {
		httpx.WriteJSONError(fmt.Errorf("update profile: %w", httpx.ErrBadRequest), w)
		return
	}

	var req UpdateProfileRequest
	if err := httpx.DecodeJSON(&req, r); err != nil {
		httpx.WriteJSONError(err, w)
		return
	}

	claims, ok := middleware.GetUserClaims(r)
	if !ok {
		httpx.WriteJSONError(httpx.ErrUnauthorized, w)
		return
	}
	if claims.UserID.String() != userID {
		httpx.WriteJSONError(httpx.ErrForbidden, w)
		return
	}
	req.UserID = claims.UserID

	if err := h.service.UpdateProfile(r.Context(), claims.UserID, req); err != nil {
		httpx.WriteJSONError(fmt.Errorf("update profile: %w", err), w)
		return
	}

	httpx.WriteJSON(w, 200, "Updated successfully")
}

func (h *ProfileHandler) GenerateAvatarUploadURL(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")

	if !httpx.IsValidUUID(userID) {
		httpx.WriteJSONError(fmt.Errorf("avatar upload url: %w", httpx.ErrBadRequest), w)
		return
	}

	claims, ok := middleware.GetUserClaims(r)
	if !ok {
		httpx.WriteJSONError(httpx.ErrUnauthorized, w)
		return
	}
	if claims.UserID.String() != userID {
		httpx.WriteJSONError(httpx.ErrForbidden, w)
		return
	}

	uploadURL, err := h.service.GenerateAvatarUploadURL(r.Context(), claims.UserID)
	if err != nil {
		httpx.WriteJSONError(fmt.Errorf("generate avatar upload url: %w", err), w)
		return
	}

	httpx.WriteJSON(w, 200, UploadURLResponse{UploadURL: uploadURL})
}

func (h *ProfileHandler) ConfirmAvatarUpload(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")

	if !httpx.IsValidUUID(userID) {
		httpx.WriteJSONError(fmt.Errorf("confirm avatar: %w", httpx.ErrBadRequest), w)
		return
	}

	claims, ok := middleware.GetUserClaims(r)
	if !ok {
		httpx.WriteJSONError(httpx.ErrUnauthorized, w)
		return
	}
	if claims.UserID.String() != userID {
		httpx.WriteJSONError(httpx.ErrForbidden, w)
		return
	}

	resp, err := h.service.ConfirmAvatarUpload(r.Context(), claims.UserID)
	if err != nil {
		httpx.WriteJSONError(fmt.Errorf("confirm avatar: %w", err), w)
		return
	}

	httpx.WriteJSON(w, 200, resp)
}

// Health
func Health(db *pgxpool.Pool, imgStorage *ImageStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := http.StatusOK
		status := "ok"

		postgresStatus := "healthy"
		if err := db.Ping(r.Context()); err != nil {
			code = http.StatusServiceUnavailable
			postgresStatus = "unhealthy"
			status = "unhealthy"
		}

		garageStatus := "healthy"
		if err := imgStorage.Ping(r.Context()); err != nil {
			code = http.StatusServiceUnavailable
			garageStatus = "unhealthy"
			status = "unhealthy"
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		json.NewEncoder(w).Encode(map[string]any{
			"service":    "profile-service",
			"status":     status,
			"PostgreSQL": postgresStatus,
			"Garage":     garageStatus,
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
		})
	}
}
