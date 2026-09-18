package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/ArtemYarin/pinterest-clone-api/pkg/postgres"
	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
	if err := godotenv.Load(".env.dev"); err != nil {
		log.Println("file .env.dev not found, using system env vars")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	poolCfg := postgres.PoolConfig{
		MaxConns:          5,
		MinConns:          1,
		MaxConnIdleTime:   2 * time.Minute,
		MaxConnLifetime:   5 * time.Minute,
		HealthCheckPeriod: 30 * time.Second,
	}

	authPool, err := postgres.NewPool(authPostgresDSN(), poolCfg)
	if err != nil {
		log.Fatalf("connecting to auth db: %v", err)
	}
	defer authPool.Close()

	pinPool, err := postgres.NewPool(pinPostgresDSN(), poolCfg)
	if err != nil {
		log.Fatalf("connecting to pin db: %v", err)
	}
	defer pinPool.Close()

	useSSL, _ := strconv.ParseBool(os.Getenv("MINIO_USE_SSL"))
	minioClient, err := minio.New(os.Getenv("MINIO_ENDPOINT"), &minio.Options{
		Creds:  credentials.NewStaticV4(os.Getenv("MINIO_USER"), os.Getenv("MINIO_PASSWORD"), ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatalf("creating minio client: %v", err)
	}

	bucket := os.Getenv("MINIO_BUCKET")
	if err := ensureBucket(ctx, minioClient, bucket); err != nil {
		log.Fatalf("ensuring minio bucket: %v", err)
	}

	email := getEnvDefault("SEED_USER_EMAIL", "seed@example.com")
	password := getEnvDefault("SEED_USER_PASSWORD", "SeedPassword123!")

	userID, err := ensureSeedUser(ctx, authPool, email, password)
	if err != nil {
		log.Fatalf("seeding user: %v", err)
	}
	log.Printf("seed user ready: %s (%s)", email, userID)

	imgDir := getEnvDefault("SEED_IMG_DIR", "seed/img")
	inserted, skipped, err := seedPins(ctx, pinPool, minioClient, bucket, userID, imgDir)
	if err != nil {
		log.Fatalf("seeding pins: %v", err)
	}
	log.Printf("seeding complete: %d pins created, %d already present", inserted, skipped)
}

func authPostgresDSN() string {
	return buildDSN("AUTH_POSTGRES_USER", "AUTH_POSTGRES_PASSWORD", "AUTH_POSTGRES_HOST", "AUTH_POSTGRES_PORT", "AUTH_POSTGRES_DB")
}

func pinPostgresDSN() string {
	return buildDSN("PIN_POSTGRES_USER", "PIN_POSTGRES_PASSWORD", "PIN_POSTGRES_HOST", "PIN_POSTGRES_PORT", "PIN_POSTGRES_DB")
}

func buildDSN(userVar, passVar, hostVar, portVar, dbVar string) string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv(userVar), os.Getenv(passVar), os.Getenv(hostVar), os.Getenv(portVar), os.Getenv(dbVar))
}

func getEnvDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
