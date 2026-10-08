package main

import (
	"context"
	"errors"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	likesv1 "github.com/ArtemYarin/pinterest-clone-api/gen/likes/v1"
	pinv1 "github.com/ArtemYarin/pinterest-clone-api/gen/pin/v1"
	"github.com/ArtemYarin/pinterest-clone-api/handlers/users"
	"github.com/ArtemYarin/pinterest-clone-api/pkg/middleware"
	"github.com/ArtemYarin/pinterest-clone-api/router"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Load .env file
	if err := godotenv.Load(".env.dev"); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			log.Println("file .env.dev not found, using system env vars")
		} else {
			log.Fatalf("failed to parse .env.dev: %v", err)
		}
	}

	// Rate Limiter
	rateLimiter := middleware.IPRateLimiter{
		Buckets:  make(map[string]*middleware.TokenBucket),
		Rate:     5,
		Capacity: 10,
	}

	// gRPC clients (connections are established lazily on first call)
	interactionConn, err := grpc.NewClient(
		net.JoinHostPort(os.Getenv("INTERACTION_HOST"), os.Getenv("INTERACTION_GRPC_PORT")),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to create interaction gRPC client: %v", err)
	}
	defer interactionConn.Close()

	pinConn, err := grpc.NewClient(
		net.JoinHostPort(os.Getenv("PIN_HOST"), os.Getenv("PIN_GRPC_PORT")),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to create pin gRPC client: %v", err)
	}
	defer pinConn.Close()

	// Wiring
	likedPinsHandler := users.NewLikedPinsHandler(
		likesv1.NewLikesServiceClient(interactionConn),
		pinv1.NewPinServiceClient(pinConn))

	r := router.SetupRouter(&rateLimiter, likedPinsHandler)

	// Server setup
	port := os.Getenv("GATEWAY_PORT")
	srv := http.Server{
		Addr:           ":" + port,
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
