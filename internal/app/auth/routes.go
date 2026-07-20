package auth

import (
	"github.com/ArtemYarin/pinterest-clone-api/internal/middleware"
	"github.com/go-chi/chi/v5"
)

func UserRouter(handler *UserHandler) chi.Router {
	r := chi.NewRouter()

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
