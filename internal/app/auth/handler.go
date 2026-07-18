package auth

import (
	"encoding/json"
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
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		json.NewEncoder(w).Encode(map[string]string{
			"code":    "400",
			"message": "Can't decode json body",
		})
		return
	}

	// Service call
	resp, err := h.service.RegisterUser(r.Context(), user)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{
			"code":    "500",
			"message": err.Error(),
		})
		return
	}

	// Response writing
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(resp)
}

func (h *UserHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	// Decode body
	var user CredentialsUserRequest
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		json.NewEncoder(w).Encode(map[string]string{
			"code":    "400",
			"message": "Can't decode json body",
		})
		return
	}

	// Service call
	token, err := h.service.LoginUser(r.Context(), user)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{
			"code":    "500",
			"message": err.Error(),
		})
		return
	}

	// Response writing
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func (h *UserHandler) GetUserByEmail(w http.ResponseWriter, r *http.Request) {
	email := chi.URLParam(r, "email")

	// Service call
	resp, err := h.service.GetUserByEmail(r.Context(), email)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{
			"code":    "500",
			"message": err.Error(),
		})
		return
	}

	// Response writing
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(resp)
}

func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Service call
	resp, err := h.service.GetUserByID(r.Context(), id)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{
			"code":    "500",
			"message": err.Error(),
		})
		return
	}

	// Response writing
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(resp)
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	// Get id from url parameters
	id := chi.URLParam(r, "id")

	// Decode json
	var user UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		json.NewEncoder(w).Encode(map[string]string{
			"code":    "400",
			"message": "Can't decode json body",
		})
		return
	}

	// Security
	claims, ok := middleware.GetUserClaims(r)
	if !ok {
		json.NewEncoder(w).Encode(map[string]string{
			"code":    "401",
			"message": "Unauthorized",
		})
		return
	}

	// Compare provided id with JWT claims
	if claims.UserID.String() != id {
		json.NewEncoder(w).Encode(map[string]string{
			"code":    "403",
			"message": "Forbidden",
		})
		return
	}

	user.Id = claims.UserID

	// Service call
	err := h.service.UpdateUser(r.Context(), user)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{
			"code":    "500",
			"message": err.Error(),
		})
	}

	// Response writing
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode("Updated successfully")
}
