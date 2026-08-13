package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ArtemYarin/pinterest-clone-api/pkg/middleware"
	"github.com/ArtemYarin/pinterest-clone-api/router"
)

func main() {
	// Rate Limiter
	rateLimiter := middleware.IPRateLimiter{
		Buckets:  make(map[string]*middleware.TokenBucket),
		Rate:     5,
		Capacity: 10,
	}

	// Wiring
	r := router.SetupRouter(&rateLimiter)

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
