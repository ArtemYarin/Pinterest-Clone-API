package router

import (
	"github.com/ArtemYarin/pinterest-clone-api/internal/app/auth"
	"github.com/ArtemYarin/pinterest-clone-api/internal/app/pin"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func SetupRouter(userHandler auth.UserHandler, pinHandler pin.PinHandler, authPool *pgxpool.Pool, pinPool *pgxpool.Pool) chi.Router {
	r := chi.NewRouter()

	r.Mount("/auth", auth.UserRouter(&userHandler, authPool))
	r.Mount("/pins", pin.PinRouter(&pinHandler, pinPool))

	return r
}
