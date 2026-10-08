package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	likesv1 "github.com/ArtemYarin/pinterest-clone-api/gen/likes/v1"
	"github.com/ArtemYarin/pinterest-clone-api/pkg/jwt"
	"github.com/ArtemYarin/pinterest-clone-api/pkg/postgres"
	"github.com/ArtemYarin/pinterest-clone-api/services/interaction-service/internal/likes"
	likesgrpc "github.com/ArtemYarin/pinterest-clone-api/services/interaction-service/internal/likes/grpc"
	"github.com/ArtemYarin/pinterest-clone-api/services/interaction-service/internal/shared/db"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

func main() {
	// Load .env file
	if err := godotenv.Load(".env.dev"); err != nil {
		log.Println("file .env.dev not found, using system env vars")
	}

	// JWT
	jm, err := jwt.NewManager(os.Getenv("JWT_SECRET"), time.Hour)
	if err != nil {
		log.Fatalf("Failed to init JWT: %v", err)
	}

	// Postgres
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

	// Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
	})
	log.Println("Connected to Redis successfully")

	// Flush worker
	w := likes.NewWorker(pool, redisClient, 30*time.Second)
	w.Start()
	defer w.Stop()
	log.Println("Started background worker successfully")

	// Wiring
	likeRepo := likes.NewLikeRepository(pool)
	likeService := likes.NewLikeService(likeRepo, redisClient)
	likeHandler := likes.NewLikeHandler(likeService)

	r := likes.LikeRouter(&likeHandler, pool, jm)

	// Server setup
	srv := http.Server{
		Addr:           ":" + os.Getenv("INTERACTION_PORT"),
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

	// gRPC server (internal calls from gateway)
	lis, err := net.Listen("tcp", ":"+os.Getenv("INTERACTION_GRPC_PORT"))
	if err != nil {
		log.Fatalf("Failed to listen for gRPC: %v", err)
	}
	grpcServer := grpc.NewServer()
	likesv1.RegisterLikesServiceServer(grpcServer, likesgrpc.NewServer(likeService))
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()
	log.Println("Started gRPC server successfully")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	grpcServer.GracefulStop()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("forced shutdown:", err)
	}
	log.Println("server stopped cleanly")
}
