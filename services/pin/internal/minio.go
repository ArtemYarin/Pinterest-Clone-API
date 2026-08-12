package pin

import (
	"context"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type ImageStorage struct {
	client *minio.Client
	bucket string
}

func NewImageStorage(endpoint, accessKey, secretKey, bucket string) (*ImageStorage, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, err
	}

	return &ImageStorage{client: client, bucket: bucket}, nil
}

func (s *ImageStorage) GenerateUploadURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	url, err := s.client.PresignedPutObject(ctx, s.bucket, objectKey, expiry)
	if err != nil {
		return "", err
	}
	return url.String(), nil
}

func (s *ImageStorage) GenerateDownloadURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	// Explicit param to display image (not download)
	reqParams := make(url.Values)
	reqParams.Set("response-content-disposition", "inline")

	url, err := s.client.PresignedGetObject(ctx, s.bucket, objectKey, expiry, reqParams)
	if err != nil {
		return "", err
	}
	return url.String(), nil
}
