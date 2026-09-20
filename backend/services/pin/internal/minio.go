package pin

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type ImageStorage struct {
	client        *minio.Client
	presignClient *minio.Client
	bucket        string
}

func NewImageStorage(ctx context.Context, internalEndpoint, publicEndpoint, accessKey, secretKey, bucket string, useSSL bool) (*ImageStorage, error) {
	creds := credentials.NewStaticV4(accessKey, secretKey, "")

	// Client for real backend<->Garage traffic, uses Docker network
	client, err := minio.New(internalEndpoint, &minio.Options{
		Creds:  creds,
		Secure: useSSL,
		Region: "garage",
	})
	if err != nil {
		return nil, err
	}

	// Client used ONLY to build presigned URLs used outside Docker
	presignClient, err := minio.New(publicEndpoint, &minio.Options{
		Creds:  creds,
		Secure: useSSL,
		Region: "garage",
	})
	if err != nil {
		return nil, err
	}

	// Initialize bucket if doesn't exist
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("check Garage bucket exists: %w", err)
	}
	if !exists {
		err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("initialize Garage bucket: %w", err)
		}
	}

	return &ImageStorage{client: client, presignClient: presignClient, bucket: bucket}, nil
}

func (s *ImageStorage) GenerateUploadURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	url, err := s.presignClient.PresignedPutObject(ctx, s.bucket, objectKey, expiry)
	if err != nil {
		return "", err
	}
	return url.String(), nil
}

func (s *ImageStorage) GenerateDownloadURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	// Explicit param to display image (not download)
	reqParams := make(url.Values)
	reqParams.Set("response-content-disposition", "inline")

	url, err := s.presignClient.PresignedGetObject(ctx, s.bucket, objectKey, expiry, reqParams)
	if err != nil {
		return "", err
	}
	return url.String(), nil
}

func (s *ImageStorage) StatObject(ctx context.Context, objectKey string) (minio.ObjectInfo, error) {
	return s.client.StatObject(ctx, s.bucket, objectKey, minio.StatObjectOptions{})
}

func (s *ImageStorage) RemoveObject(ctx context.Context, objectKey string) error {
	return s.client.RemoveObject(ctx, s.bucket, objectKey, minio.RemoveObjectOptions{})
}

func (s *ImageStorage) Ping(ctx context.Context) error {
	_, err := s.client.BucketExists(ctx, s.bucket)
	return err
}
