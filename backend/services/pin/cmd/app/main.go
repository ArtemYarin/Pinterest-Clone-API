package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	pinv1 "github.com/ArtemYarin/pinterest-clone-api/gen/pin/v1"
	"github.com/ArtemYarin/pinterest-clone-api/pkg/jwt"
	"github.com/ArtemYarin/pinterest-clone-api/pkg/postgres"
	pin "github.com/ArtemYarin/pinterest-clone-api/services/pin-service/internal"
	pingrpc "github.com/ArtemYarin/pinterest-clone-api/services/pin-service/internal/grpc"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
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

	// Connecting to db
	dbUrl := pin.GetPinPostgresDSN()
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

	// Validator
	validate := validator.New()

	// MiniO image storage
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	useSSL, _ := strconv.ParseBool(os.Getenv("GARAGE_USE_SSL"))
	miniO, err := pin.NewImageStorage(
		ctx,
		os.Getenv("GARAGE_INTERNAL_ENDPOINT"),
		os.Getenv("GARAGE_PUBLIC_ENDPOINT"),
		os.Getenv("GARAGE_USER"),
		os.Getenv("GARAGE_PASSWORD"),
		os.Getenv("GARAGE_BUCKET"),
		useSSL)
	if err != nil {
		log.Fatalf("Failed to connect to Garage: %v", err)
	}
	log.Println("Connected to Garage successfully")

	// Browsers upload images straight to Garage, which needs bucket CORS rules.
	corsOrigins := []string{"*"}
	if v := os.Getenv("GARAGE_CORS_ORIGINS"); v != "" {
		corsOrigins = strings.Split(v, ",")
	}
	if err := miniO.AllowBrowserAccess(ctx, corsOrigins); err != nil {
		log.Fatalf("Failed to set Garage bucket CORS: %v", err)
	}

	// Wiring
	pinRepo := pin.NewPinRepository(pool)

	// Background cleanup worker for stale pending uploads
	cleanupWorker := pin.NewCleanupWorker(
		pinRepo, miniO,
		pin.GetDurationEnv("PIN_CLEANUP_INTERVAL", 10*time.Minute),
		pin.GetDurationEnv("PIN_CLEANUP_TTL", 60*time.Minute),
		pin.GetIntEnv("PIN_CLEANUP_BATCH_SIZE", 100),
	)
	cleanupWorker.Start()
	defer cleanupWorker.Stop()
	log.Println("Started pin cleanup worker successfully")

	pinService := pin.NewPinService(pinRepo, validate, miniO)
	pinHandler := pin.NewPinHandler(pinService)

	r := pin.PinRouter(&pinHandler, pool, miniO, jm)

	// Server setup
	srv := http.Server{
		Addr:           ":" + os.Getenv("PIN_PORT"),
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
	lis, err := net.Listen("tcp", ":"+os.Getenv("PIN_GRPC_PORT"))
	if err != nil {
		log.Fatalf("Failed to listen for gRPC: %v", err)
	}
	grpcServer := grpc.NewServer()
	pinv1.RegisterPinServiceServer(grpcServer, pingrpc.NewServer(pinService))
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

	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("forced shutdown:", err)
	}
	log.Println("server stopped cleanly")
}
