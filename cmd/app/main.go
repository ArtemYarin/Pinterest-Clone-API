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

	"github.com/ArtemYarin/pinterest-clone-api/internal/app/auth"
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

	// Users db
	dbUrl := postgres.GetUsersPostgresDSN()
	config := postgres.PoolConfig{
		MaxConns:          25,
		MinConns:          10,
		MaxConnIdleTime:   2 * time.Minute,
		MaxConnLifetime:   5 * time.Minute,
		HealthCheckPeriod: 30 * time.Second,
	}

	usersPool, err := postgres.NewPool(dbUrl, config)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer usersPool.Close()
	log.Println("Connected to users PostgreSQL successfully")

	// Pins db
	dbUrl = postgres.GetPinsPostgresDSN()
	config = postgres.PoolConfig{
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
	userRepo := auth.NewUserRepository(usersPool)
	userService := auth.NewUserService(userRepo, validate)
	userHandler := auth.NewUserHandler(userService)

	pinRepo := pin.NewPinRepository(pinsPool)
	pinService := pin.NewPinService(pinRepo, validate)
	pinHandler := pin.NewPinHandler(pinService)

	r := router.SetupRouter(userHandler, pinHandler, usersPool, pinsPool)

	r.Get("/health", healthHandler(usersPool, pinsPool))

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

func healthHandler(authDb *pgxpool.Pool, pinDb *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := http.StatusOK
		status := "ok"
		postgresAuthStatus := "healthy"
		postgresPinStatus := "healthy"
		if err := authDb.Ping(r.Context()); err != nil {
			code = http.StatusServiceUnavailable
			postgresAuthStatus = "unhealthy"
			status = "unhealthy"
		}
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
				"PostgreSQL Auth":     postgresAuthStatus,
				"PostgreSQL Pin":      postgresPinStatus,
			},
		})
	}
}
