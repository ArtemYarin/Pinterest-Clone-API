package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ArtemYarin/pinterest-clone-api/services/auth-service"
	"github.com/ArtemYarin/pinterest-clone-api/services/auth-service/postgres"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("file .env not found, using system env vars")
	}

	// Connecting to db
	dbUrl := postgres.GetAuthPostgresDSN()
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
	log.Println("Connected to auth PostgreSQL successfully")

	// Validator
	validate := validator.New()

	// Wiring
	userRepo := auth.NewUserRepository(pool)
	userService := auth.NewUserService(userRepo, validate)
	userHandler := auth.NewUserHandler(userService)

	r := auth.UserRouter(&userHandler, pool)

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
