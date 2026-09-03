package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ArtemYarin/pinterest-clone-api/pkg/postgres"
	"github.com/ArtemYarin/pinterest-clone-api/services/interaction-service/internal/likes"
	"github.com/ArtemYarin/pinterest-clone-api/services/interaction-service/internal/shared/db"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("file .env not found, using system env vars")
	}

	// Connecting to db
	dbUrl := db.GetInteractionPostgresDSN()
	config := postgres.PoolConfig{
		MaxConns:          25,
		MinConns:          10,
		MaxConnIdleTime:   2 * time.Minute,
		MaxConnLifetime:   5 * time.Minute,
		HealthCheckPeriod: 30 * time.Second,
	}
	pool, err := postgres.NewPool(dbUrl, config)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer pool.Close()
	log.Println("Connected to PostgreSQL successfully")

	// Wiring
	likeRepo := likes.NewLikeRepository(pool)
	likeService := likes.NewLikeService(likeRepo)
	likeHandler := likes.NewLikeHandler(likeService)

	r := likes.LikeRouter(&likeHandler, pool)

	// Server setup
	srv := http.Server{
		Addr:           ":8083",
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
