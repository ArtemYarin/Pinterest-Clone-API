package router

import (
	"github.com/ArtemYarin/pinterest-clone-api/internal/app/pin"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func SetupRouter(pinHandler pin.PinHandler, pinPool *pgxpool.Pool) chi.Router {
	r := chi.NewRouter()

	r.Mount("/pins", pin.PinRouter(&pinHandler, pinPool))

	return r
}
