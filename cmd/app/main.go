package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ArtemYarin/pinterest-clone-api/internal/app/pin"
	"github.com/ArtemYarin/pinterest-clone-api/internal/postgres"
	"github.com/ArtemYarin/pinterest-clone-api/internal/router"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("file .env not found, using system env vars")
	}

	// Connecting to db

	// Pins db
	dbUrl := postgres.GetPinsPostgresDSN()
	config := postgres.PoolConfig{
		MaxConns:          25,
		MinConns:          10,
		MaxConnIdleTime:   2 * time.Minute,
		MaxConnLifetime:   5 * time.Minute,
		HealthCheckPeriod: 30 * time.Second,
	}

	pinsPool, err := postgres.NewPool(dbUrl, config)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer pinsPool.Close()
	log.Println("Connected to pins PostgreSQL successfully")

	// Validator
	validate := validator.New()

	// Wiring
	pinRepo := pin.NewPinRepository(pinsPool)
	pinService := pin.NewPinService(pinRepo, validate)
	pinHandler := pin.NewPinHandler(pinService)

	r := router.SetupRouter(pinHandler, pinsPool)

	r.Get("/health", healthHandler(pinsPool))

	// Server setup
	srv := http.Server{
		Addr:           ":8080",
		Handler:        r,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// Clear shutdown
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("forced shutdown:", err)
	}
	log.Println("server stopped cleanly")
}

func healthHandler(pinDb *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := http.StatusOK
		status := "ok"
		postgresPinStatus := "healthy"
		if err := pinDb.Ping(r.Context()); err != nil {
			code = http.StatusServiceUnavailable
			postgresPinStatus = "unhealthy"
			status = "unhealthy"
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		json.NewEncoder(w).Encode(map[string]any{
			"status":    status,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"services": map[string]string{
				"pinterest-clone-api": "ok",
				"PostgreSQL Pin":      postgresPinStatus,
			},
		})
	}
}
