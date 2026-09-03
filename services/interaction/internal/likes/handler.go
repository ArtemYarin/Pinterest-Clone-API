package likes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ArtemYarin/pinterest-clone-api/pkg/middleware"
	"github.com/ArtemYarin/pinterest-clone-api/services/interaction-service/internal/shared/errs"
	"github.com/ArtemYarin/pinterest-clone-api/services/interaction-service/internal/shared/jsonutil"
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
		jsonutil.WriteJSONError(fmt.Errorf("add like: %w", errs.ErrBadRequest), w)
		return
	}

	claims, ok := middleware.GetUserClaims(r)
	if !ok {
		jsonutil.WriteJSONError(fmt.Errorf("add like: %w", errs.ErrUnauthorized), w)
		return
	}

	if err := h.service.AddLike(r.Context(), claims.UserID, pinID); err != nil {
		jsonutil.WriteJSONError(fmt.Errorf("add like: %w", err), w)
		return
	}

	jsonutil.WriteJSON(w, http.StatusOK, "Liked successfully")
}

func (h *LikeHandler) RemoveLike(w http.ResponseWriter, r *http.Request) {
	pinID, err := uuid.Parse(chi.URLParam(r, "pinID"))
	if err != nil {
		jsonutil.WriteJSONError(fmt.Errorf("remove like: %w", errs.ErrBadRequest), w)
		return
	}

	claims, ok := middleware.GetUserClaims(r)
	if !ok {
		jsonutil.WriteJSONError(fmt.Errorf("remove like: %w", errs.ErrUnauthorized), w)
		return
	}

	if err := h.service.RemoveLike(r.Context(), claims.UserID, pinID); err != nil {
		jsonutil.WriteJSONError(fmt.Errorf("remove like: %w", err), w)
		return
	}

	jsonutil.WriteJSON(w, http.StatusOK, "Unliked successfully")
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
