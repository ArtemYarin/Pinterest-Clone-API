package likes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ArtemYarin/pinterest-clone-api/pkg/httpx"
	"github.com/ArtemYarin/pinterest-clone-api/pkg/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LikeHandler struct {
	service LikeService
}

func NewLikeHandler(s LikeService) LikeHandler {
	return LikeHandler{service: s}
}

func (h *LikeHandler) AddLike(w http.ResponseWriter, r *http.Request) {
	pinID, err := uuid.Parse(chi.URLParam(r, "pinID"))
	if err != nil {
		httpx.WriteJSONError(fmt.Errorf("add like: %w", httpx.ErrBadRequest), w)
		return
	}

	claims, ok := middleware.GetUserClaims(r)
	if !ok {
		httpx.WriteJSONError(fmt.Errorf("add like: %w", httpx.ErrUnauthorized), w)
		return
	}

	if err := h.service.AddLike(r.Context(), claims.UserID, pinID); err != nil {
		httpx.WriteJSONError(fmt.Errorf("add like: %w", err), w)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, "Liked successfully")
}

func (h *LikeHandler) RemoveLike(w http.ResponseWriter, r *http.Request) {
	pinID, err := uuid.Parse(chi.URLParam(r, "pinID"))
	if err != nil {
		httpx.WriteJSONError(fmt.Errorf("remove like: %w", httpx.ErrBadRequest), w)
		return
	}

	claims, ok := middleware.GetUserClaims(r)
	if !ok {
		httpx.WriteJSONError(fmt.Errorf("remove like: %w", httpx.ErrUnauthorized), w)
		return
	}

	if err := h.service.RemoveLike(r.Context(), claims.UserID, pinID); err != nil {
		httpx.WriteJSONError(fmt.Errorf("remove like: %w", err), w)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, "Unliked successfully")
}

func (h *LikeHandler) GetLikeCount(w http.ResponseWriter, r *http.Request) {
	pinID, err := uuid.Parse(chi.URLParam(r, "pinID"))
	if err != nil {
		httpx.WriteJSONError(fmt.Errorf("get like count: %w", httpx.ErrBadRequest), w)
		return
	}

	count, err := h.service.GetLikeCount(r.Context(), pinID)
	if err != nil {
		httpx.WriteJSONError(fmt.Errorf("get like count: %w", err), w)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"pin_id": pinID,
		"count":  *count,
	})
}

func (h *LikeHandler) HasLiked(w http.ResponseWriter, r *http.Request) {
	pinID, err := uuid.Parse(chi.URLParam(r, "pinID"))
	if err != nil {
		httpx.WriteJSONError(fmt.Errorf("has liked: %w", httpx.ErrBadRequest), w)
		return
	}

	claims, ok := middleware.GetUserClaims(r)
	if !ok {
		httpx.WriteJSONError(fmt.Errorf("has liked: %w", httpx.ErrUnauthorized), w)
		return
	}

	liked, err := h.service.HasLiked(r.Context(), claims.UserID, pinID)
	if err != nil {
		httpx.WriteJSONError(fmt.Errorf("has liked: %w", err), w)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"pin_id": pinID,
		"liked":  *liked,
	})
}

func (h *LikeHandler) HasLikedBatch(w http.ResponseWriter, r *http.Request) {
	rawIDs := strings.Split(r.URL.Query().Get("pin_ids"), ",")

	pinIDs := make(uuid.UUIDs, 0, len(rawIDs))
	for _, rawID := range rawIDs {
		rawID = strings.TrimSpace(rawID)
		if rawID == "" {
			continue
		}
		pinID, err := uuid.Parse(rawID)
		if err != nil {
			httpx.WriteJSONError(fmt.Errorf("has liked batch: %w", httpx.ErrBadRequest), w)
			return
		}
		pinIDs = append(pinIDs, pinID)
	}

	if len(pinIDs) == 0 {
		httpx.WriteJSONError(fmt.Errorf("has liked batch: %w", httpx.ErrBadRequest), w)
		return
	}

	claims, ok := middleware.GetUserClaims(r)
	if !ok {
		httpx.WriteJSONError(fmt.Errorf("has liked batch: %w", httpx.ErrUnauthorized), w)
		return
	}

	liked, err := h.service.HasLikedBatch(r.Context(), claims.UserID, pinIDs)
	if err != nil {
		httpx.WriteJSONError(fmt.Errorf("has liked batch: %w", err), w)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, liked)
}

func (h *LikeHandler) ListLikedByUser(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserClaims(r)
	if !ok {
		httpx.WriteJSONError(fmt.Errorf("list liked by user: %w", httpx.ErrUnauthorized), w)
		return
	}

	pinIDs, err := h.service.ListLikedByUser(r.Context(), claims.UserID)
	if err != nil {
		httpx.WriteJSONError(fmt.Errorf("list liked by user: %w", err), w)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"pin_ids": pinIDs,
	})
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
			"service":    "interaction-service",
			"status":     status,
			"PostgreSQL": postgresStatus,
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
		})
	}
}
