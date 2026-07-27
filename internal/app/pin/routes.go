package pin

import (
	"github.com/ArtemYarin/pinterest-clone-api/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func PinRouter(handler *PinHandler, pinPool *pgxpool.Pool) chi.Router {
	r := chi.NewRouter()

	r.Get("/health", Health(pinPool))

	r.Get("/{id}", handler.GetPin)

	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware)

		r.Post("/", handler.CreatePin)
		r.Patch("/{id}", handler.UpdatePin)
		r.Delete("/{id}", handler.DeletePin)
	})
	return r
}
