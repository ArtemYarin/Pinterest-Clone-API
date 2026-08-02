package auth

import (
	"github.com/ArtemYarin/pinterest-clone-api/services/auth-service/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func UserRouter(handler *UserHandler, userPool *pgxpool.Pool) chi.Router {
	r := chi.NewRouter()

	r.Get("/health", Health(userPool))

	r.Post("/signup", handler.RegisterUser)
	r.Post("/login", handler.LoginUser)
	r.Get("/{id}", handler.GetUserByID)
	r.Get("/email/{email}", handler.GetUserByEmail)

	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware)

		r.Patch("/{id}", handler.UpdateUser)
	})
	return r
}
