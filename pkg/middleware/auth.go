package middleware

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/ArtemYarin/pinterest-clone-api/pkg/jwt"
)

type contextKey string

const UserKey contextKey = "user"

func GetUserClaims(r *http.Request) (*jwt.Claims, bool) {
	userClaims, ok := r.Context().Value(UserKey).(*jwt.Claims)
	return userClaims, ok
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extraction
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			log.Printf("auth Error: missing or invalid Authorization header")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Missing or invalid Authorization header",
			})
			return
		}
		raw := strings.TrimPrefix(authHeader, "Bearer ")

		// Validation
		claims, err := jwt.ValidateToken(raw)
		if err != nil {
			log.Printf("auth error: invalid or expired token")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"code":    "401",
				"message": "Invalid or expired token",
			})
			return
		}

		// Attach user and pass
		ctx := context.WithValue(r.Context(), UserKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
