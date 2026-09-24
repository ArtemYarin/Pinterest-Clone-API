package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/ArtemYarin/pinterest-clone-api/pkg/postgres"
	"github.com/ArtemYarin/pinterest-clone-api/pkg/rabbitmq"
	profile "github.com/ArtemYarin/pinterest-clone-api/services/profile-service/internal"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	rmq "github.com/rabbitmq/rabbitmq-amqp-go-client/pkg/rabbitmqamqp"
)

func main() {
	// Load .env file
	if err := godotenv.Load(".env.dev"); err != nil {
		log.Println("file .env.dev not found, using system env vars")
	}

	// Connecting to db
	dbUrl := profile.GetProfilePostgresDSN()
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
	log.Println("Connected to profile PostgreSQL successfully")

	// Validator
	validate := validator.New()

	// Garage image storage (avatars)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	useSSL, _ := strconv.ParseBool(os.Getenv("GARAGE_USE_SSL"))
	imgStorage, err := profile.NewImageStorage(
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

	// Wiring
	profileRepo := profile.NewProfileRepository(pool)
	profileService := profile.NewProfileService(profileRepo, validate, imgStorage)
	profileHandler := profile.NewProfileHandler(profileService)

	r := profile.ProfileRouter(&profileHandler, pool, imgStorage)

	// Consumer for auth-service's user.registered events
	consumerCtx, consumerCancel := context.WithCancel(context.Background())
	env := rmq.NewEnvironment(os.Getenv("RABBITMQ_URL"), nil)
	conn, err := env.NewConnection(consumerCtx)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer func() {
		_ = env.CloseConnections(context.Background())
	}()

	_, err = conn.Management().DeclareQueue(consumerCtx, &rmq.QuorumQueueSpecification{Name: "userCreated"})
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	consumer, err := conn.NewConsumer(consumerCtx, "userCreated", nil)
	if err != nil {
		log.Panicf("Failed to create consumer: %v", err)
	}
	defer func() { _ = consumer.Close(context.Background()) }()
	log.Printf("Waiting for messages.")

	// Receive loop
	// WaitGroup makes sure that loop can fully exit before resources get closed
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			delivery, err := consumer.Receive(consumerCtx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return
				}
				log.Fatalf("Failed to receive a message: %v", err)
			}

			settleCtx, settleCancel := context.WithTimeout(context.Background(), 5*time.Second)
			msg := delivery.Message()
			var evt rabbitmq.UserRegisteredEvent
			if len(msg.Data) == 0 || json.Unmarshal(msg.Data[0], &evt) != nil || evt.UserID == uuid.Nil {
				log.Print("malformed user.registered event, discarding")
				_ = delivery.Discard(settleCtx, nil)
				settleCancel()
				continue
			}

			tCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			_, err = profileService.CreateFromRegistration(tCtx, evt.UserID, evt.SuggestedUsername)
			cancel()

			if err != nil {
				log.Printf("create profile (user %s): %v, requeueing", evt.UserID, err)
				_ = delivery.Requeue(settleCtx)
			} else if err := delivery.Accept(settleCtx); err != nil {
				log.Printf("accept failed: %v", err)
			}
			settleCancel()
		}
	}()

	// Server setup
	srv := http.Server{
		Addr:           ":" + os.Getenv("PROFILE_PORT"),
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

	consumerCancel()
	wg.Wait()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal("forced shutdown:", err)
	}
	log.Println("server stopped cleanly")
}
