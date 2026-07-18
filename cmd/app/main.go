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

	"github.com/ArtemYarin/pinterest-clone-api/internal/postgres"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("file .env not found, using system env vars")
	}

	// Connecting to db
	dbUrl := postgres.GetPostgresDSN()
	config := postgres.PoolConfig{
		MaxConns:          25,
		MinConns:          10,
		MaxConnIdleTime:   2 * time.Minute,
		MaxConnLifetime:   5 * time.Minute,
		HealthCheckPeriod: 30 * time.Second,
	}

	pool, err := postgres.NewPool(dbUrl, config)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer pool.Close()
	log.Println("Connected to PostgreSQL successfully")

	// Wiring
	r := chi.NewRouter()
	r.Get("/health", healthHandler(pool))

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

func healthHandler(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := "ok"
		postgresStatus := "healthy"
		if err := db.Ping(r.Context()); err != nil {
			postgresStatus = "unhealthy"
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"status":    status,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"services": map[string]string{
				"pinterest-clone-api": "ok",
				"PostgreSQL":          postgresStatus,
			},
		})
	}
}
