package profile

import (
	"github.com/ArtemYarin/pinterest-clone-api/pkg/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func ProfileRouter(handler *ProfileHandler, profilePool *pgxpool.Pool, imgStorage *ImageStorage) chi.Router {
	r := chi.NewRouter()

	r.Get("/health", Health(profilePool, imgStorage))

	r.Get("/{userId}", handler.GetProfile)

	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware)

		r.Patch("/{userId}", handler.UpdateProfile)
		r.Post("/{userId}/avatar/upload-url", handler.GenerateAvatarUploadURL)
		r.Post("/{userId}/avatar/confirm", handler.ConfirmAvatarUpload)
	})
	return r
}
