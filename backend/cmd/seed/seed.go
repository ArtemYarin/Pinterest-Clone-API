package main

import (
	"context"
	"errors"
	"fmt"
	"mime"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
	"golang.org/x/crypto/bcrypt"
)

func ensureBucket(ctx context.Context, client *minio.Client, bucket string) error {
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return fmt.Errorf("checking bucket: %w", err)
	}
	if exists {
		return nil
	}
	if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
		return fmt.Errorf("creating bucket: %w", err)
	}
	return nil
}

// ensureSeedUser returns the id of the seed user, creating it (with a
// bcrypt-hashed password, same as the auth service's signup flow) if it
// doesn't already exist.
func ensureSeedUser(ctx context.Context, pool *pgxpool.Pool, email, password string) (uuid.UUID, error) {
	var id uuid.UUID
	err := pool.QueryRow(ctx, "SELECT id FROM users WHERE email = $1", email).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, fmt.Errorf("looking up seed user: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return uuid.Nil, fmt.Errorf("hashing seed password: %w", err)
	}

	err = pool.QueryRow(ctx,
		"INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id",
		email, string(hash)).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("inserting seed user: %w", err)
	}
	return id, nil
}

// seedPins inserts each pinSeeds entry that isn't already present (matched
// on the pins.image_url unique constraint) and uploads its image to MinIO.
func seedPins(ctx context.Context, pool *pgxpool.Pool, minioClient *minio.Client, bucket string, userID uuid.UUID, imgDir string) (inserted, skipped int, err error) {
	for _, seed := range pinSeeds {
		objectKey := fmt.Sprintf("pins/%s/%s", userID, seed.Slug)

		var pinID uuid.UUID
		insertErr := pool.QueryRow(ctx,
			`INSERT INTO pins (user_id, title, image_url, description, image_status)
			 VALUES ($1, $2, $3, $4, 'confirmed')
			 ON CONFLICT (image_url) DO NOTHING
			 RETURNING id`,
			userID, seed.Title, objectKey, seed.Description).Scan(&pinID)

		if errors.Is(insertErr, pgx.ErrNoRows) {
			skipped++
			continue
		}
		if insertErr != nil {
			return inserted, skipped, fmt.Errorf("inserting pin %q: %w", seed.Title, insertErr)
		}

		imgPath := filepath.Join(imgDir, seed.Filename)
		contentType := mime.TypeByExtension(filepath.Ext(seed.Filename))
		if contentType == "" {
			contentType = "image/jpeg"
		}

		if _, err := minioClient.FPutObject(ctx, bucket, objectKey, imgPath, minio.PutObjectOptions{ContentType: contentType}); err != nil {
			return inserted, skipped, fmt.Errorf("uploading image for pin %q: %w", seed.Title, err)
		}

		inserted++
	}
	return inserted, skipped, nil
}
