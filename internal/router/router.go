package router

import (
	"github.com/ArtemYarin/pinterest-clone-api/internal/app/auth"
	"github.com/go-chi/chi/v5"
)

func SetupRouter(userHandler auth.UserHandler) chi.Router {
	r := chi.NewRouter()

	r.Mount("/auth", auth.UserRouter(&userHandler))

	return r
}
