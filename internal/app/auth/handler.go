package auth

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ArtemYarin/pinterest-clone-api/internal/middleware"
	"github.com/go-chi/chi/v5"
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
	if err := DecodeJSON(&user, r); err != nil {
		WriteJSONError(err, w)
		return
	}

	// Service call
	resp, err := h.service.RegisterUser(r.Context(), user)
	if err != nil {
		WriteJSONError(fmt.Errorf("register user: %w", err), w)
		return
	}

	// Response writing
	WriteJSON(w, http.StatusCreated, resp)
}

func (h *UserHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	// Decode body
	var user CredentialsUserRequest
	if err := DecodeJSON(&user, r); err != nil {
		WriteJSONError(err, w)
		return
	}

	// Service call
	token, err := h.service.LoginUser(r.Context(), user)
	if err != nil {
		WriteJSONError(fmt.Errorf("login user: %w", err), w)
		return
	}

	// Response writing
	WriteJSON(w, 200, map[string]string{"token": token})
}

func (h *UserHandler) GetUserByEmail(w http.ResponseWriter, r *http.Request) {
	email := chi.URLParam(r, "email")

	// Service call
	resp, err := h.service.GetUserByEmail(r.Context(), email)
	if err != nil {
		WriteJSONError(fmt.Errorf("get user by email: %w", err), w)
		return
	}

	// Response writing
	WriteJSON(w, 200, resp)
}

func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Service call
	resp, err := h.service.GetUserByID(r.Context(), id)
	if err != nil {
		WriteJSONError(fmt.Errorf("get user by id: %w", err), w)
		return
	}

	// Response writing
	WriteJSON(w, 200, resp)
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	// Get id from url parameters
	id := chi.URLParam(r, "id")

	// Decode json
	var user UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		WriteJSONError(err, w)
		return
	}

	// Security
	claims, ok := middleware.GetUserClaims(r)
	if !ok {
		WriteJSONError(errUnauthorized, w)
		return
	}

	// Compare provided id with JWT claims
	if claims.UserID.String() != id {
		WriteJSONError(errForbidden, w)
		return
	}

	user.Id = claims.UserID

	// Service call
	err := h.service.UpdateUser(r.Context(), user)
	if err != nil {
		WriteJSONError(fmt.Errorf("update user: %w", err), w)
		return
	}

	// Response writing
	WriteJSON(w, 200, "Updated successfully")
}
