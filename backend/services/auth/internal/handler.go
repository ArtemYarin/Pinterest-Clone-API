package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ArtemYarin/pinterest-clone-api/pkg/httpx"
	"github.com/ArtemYarin/pinterest-clone-api/pkg/middleware"
	"github.com/ArtemYarin/pinterest-clone-api/services/auth-service/internal/refresh"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserHandler struct {
	service UserService
}

func NewUserHandler(s UserService) UserHandler {
	return UserHandler{service: s}
}

func (h *UserHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	// Decode body
	var user CredentialsUserRequest
	if err := httpx.DecodeJSON(&user, r); err != nil {
		httpx.WriteJSONError(err, w)
		return
	}

	// Service call
	resp, err := h.service.RegisterUser(r.Context(), user)
	if err != nil {
		httpx.WriteJSONError(fmt.Errorf("register user: %w", err), w)
		return
	}

	// Refresh token (non-fatal: user can still use the access token)
	if err := h.issueRefreshToken(w, r, resp.Id); err != nil {
		log.Printf("register user: %v", err)
	}

	// Response writing
	httpx.WriteJSON(w, http.StatusCreated, resp)
}

func (h *UserHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	// Decode body
	var user CredentialsUserRequest
	if err := httpx.DecodeJSON(&user, r); err != nil {
		httpx.WriteJSONError(err, w)
		return
	}

	// Service call
	resp, err := h.service.LoginUser(r.Context(), user)
	if err != nil {
		httpx.WriteJSONError(fmt.Errorf("login user: %w", err), w)
		return
	}

	// Refresh token (non-fatal: user can still use the access token)
	if err := h.issueRefreshToken(w, r, resp.Id); err != nil {
		log.Printf("login user: %v", err)
	}

	// Response writing
	httpx.WriteJSON(w, 200, resp)
}

func (h *UserHandler) GetUserByEmail(w http.ResponseWriter, r *http.Request) {
	email := chi.URLParam(r, "email")

	// Service call
	resp, err := h.service.GetUserByEmail(r.Context(), email)
	if err != nil {
		httpx.WriteJSONError(fmt.Errorf("get user by email: %w", err), w)
		return
	}

	// Response writing
	httpx.WriteJSON(w, 200, resp)
}

func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if !httpx.IsValidUUID(id) {
		httpx.WriteJSONError(fmt.Errorf("get user by id: %w", httpx.ErrBadRequest), w)
		return
	}

	// Service call
	resp, err := h.service.GetUserByID(r.Context(), id)
	if err != nil {
		httpx.WriteJSONError(fmt.Errorf("get user by id: %w", err), w)
		return
	}

	// Response writing
	httpx.WriteJSON(w, 200, resp)
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	// Get id from url parameters
	id := chi.URLParam(r, "id")

	// Decode json
	var user UpdateUserRequest
	if err := httpx.DecodeJSON(&user, r); err != nil {
		httpx.WriteJSONError(err, w)
		return
	}

	// Security
	claims, ok := middleware.GetUserClaims(r)
	if !ok {
		httpx.WriteJSONError(httpx.ErrUnauthorized, w)
		return
	}

	// Compare provided id with JWT claims
	if claims.UserID.String() != id {
		httpx.WriteJSONError(httpx.ErrForbidden, w)
		return
	}

	user.Id = claims.UserID

	// Service call
	err := h.service.UpdateUser(r.Context(), user)
	if err != nil {
		httpx.WriteJSONError(fmt.Errorf("update user: %w", err), w)
		return
	}

	// Response writing
	httpx.WriteJSON(w, 200, "Updated successfully")
}

func (h *UserHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	// Cookie
	c, err := r.Cookie("refresh_token")
	if err != nil {
		httpx.WriteJSONError(httpx.ErrUnauthorized, w)
		return
	}

	// Find refresh token in db
	rt, err := h.service.FindRefreshToken(r.Context(), c.Value)
	if err != nil {
		if errors.Is(err, errTokenNotFound) {
			refresh.ClearRefreshCookie(w)
			httpx.WriteJSONError(fmt.Errorf("refresh: %w", httpx.ErrUnauthorized), w)
			return
		}
		httpx.WriteJSONError(fmt.Errorf("refresh: %w", err), w)
		return
	}
	if !rt.IsValid(time.Now()) {
		refresh.ClearRefreshCookie(w)
		httpx.WriteJSONError(fmt.Errorf("refresh: expired or revoked token: %w", httpx.ErrUnauthorized), w)
		return
	}

	// Revoke old token
	if err := h.service.RevokeRefreshToken(r.Context(), rt.ID); err != nil {
		if errors.Is(err, errTokenNotFound) {
			refresh.ClearRefreshCookie(w)
			httpx.WriteJSONError(fmt.Errorf("refresh: %w", httpx.ErrUnauthorized), w)
			return
		}
		httpx.WriteJSONError(fmt.Errorf("refresh: %w", err), w)
		return
	}

	// Issue new refresh token
	if err := h.issueRefreshToken(w, r, rt.UserID); err != nil {
		httpx.WriteJSONError(fmt.Errorf("refresh: %w", err), w)
		return
	}

	// Issue new access token
	token, err := h.service.IssueAccessToken(rt.UserID)
	if err != nil {
		httpx.WriteJSONError(fmt.Errorf("refresh: %w", err), w)
		return
	}

	// Response writing
	httpx.WriteJSON(w, 200, map[string]string{"token": token})
}

func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Logout is idempotent: no cookie or unknown token still logs out.
	if c, err := r.Cookie("refresh_token"); err == nil {
		err := h.service.RevokeRefreshTokenByHash(r.Context(), c.Value)
		if err != nil && !errors.Is(err, errTokenNotFound) {
			log.Printf("logout: %v", err)
		}
	}

	refresh.ClearRefreshCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

// issueRefreshToken generates a refresh token, stores its hash and sets the cookie.
// The cookie is only set if the token was saved.
func (h *UserHandler) issueRefreshToken(w http.ResponseWriter, r *http.Request, userID uuid.UUID) error {
	plain, hash, err := refresh.NewRefreshToken()
	if err != nil {
		return fmt.Errorf("generate refresh token: %w", err)
	}
	if err := h.service.SaveRefreshToken(r.Context(), userID.String(), hash); err != nil {
		return fmt.Errorf("save refresh token: %w", err)
	}
	refresh.SetRefreshCookie(w, plain)
	return nil
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
			"service":    "auth-service",
			"status":     status,
			"PostgreSQL": postgresStatus,
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
		})
	}
}
