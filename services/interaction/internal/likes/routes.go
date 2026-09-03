package likes

import (
	"github.com/ArtemYarin/pinterest-clone-api/pkg/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func LikeRouter(handler *LikeHandler, pool *pgxpool.Pool) chi.Router {
	r := chi.NewRouter()

	r.Get("/health", Health(pool))

	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware)

		r.Put("/{pinID}/like", handler.AddLike)
		r.Delete("/{pinID}/like", handler.RemoveLike)
	})
	return r
}
