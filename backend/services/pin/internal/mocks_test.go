package pin

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/mock"
)

type mockPinRepository struct {
	mock.Mock
}

func (m *mockPinRepository) CreatePin(ctx context.Context, userID uuid.UUID, imageURL string, pin CreatePinRequest) (*PinResponse, error) {
	args := m.Called(ctx, userID, imageURL, pin)
	p, _ := args.Get(0).(*PinResponse)
	return p, args.Error(1)
}

func (m *mockPinRepository) GetPinByID(ctx context.Context, id string) (*PinResponse, error) {
	args := m.Called(ctx, id)
	p, _ := args.Get(0).(*PinResponse)
	return p, args.Error(1)
}

func (m *mockPinRepository) GetPins(ctx context.Context, filters PinFilters) ([]PinResponse, int, error) {
	args := m.Called(ctx, filters)
	pins, _ := args.Get(0).([]PinResponse)
	return pins, args.Int(1), args.Error(2)
}

func (m *mockPinRepository) GetPinsByIDs(ctx context.Context, ids uuid.UUIDs) ([]PinResponse, error) {
	args := m.Called(ctx, ids)
	pins, _ := args.Get(0).([]PinResponse)
	return pins, args.Error(1)
}

func (m *mockPinRepository) UpdatePin(ctx context.Context, pin UpdatePinRequest) error {
	args := m.Called(ctx, pin)
	return args.Error(0)
}

func (m *mockPinRepository) DeletePin(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockPinRepository) UpdateImageStatus(ctx context.Context, id string, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *mockPinRepository) GetStalePendingPins(ctx context.Context, olderThan time.Time, limit int) ([]StalePin, error) {
	args := m.Called(ctx, olderThan, limit)
	stale, _ := args.Get(0).([]StalePin)
	return stale, args.Error(1)
}

type mockImageStore struct {
	mock.Mock
}

func (m *mockImageStore) GenerateUploadURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	args := m.Called(ctx, objectKey, expiry)
	return args.String(0), args.Error(1)
}

func (m *mockImageStore) GenerateDownloadURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	args := m.Called(ctx, objectKey, expiry)
	return args.String(0), args.Error(1)
}

func (m *mockImageStore) StatObject(ctx context.Context, objectKey string) (minio.ObjectInfo, error) {
	args := m.Called(ctx, objectKey)
	info, _ := args.Get(0).(minio.ObjectInfo)
	return info, args.Error(1)
}

func (m *mockImageStore) RemoveObject(ctx context.Context, objectKey string) error {
	args := m.Called(ctx, objectKey)
	return args.Error(0)
}

type mockPinService struct {
	mock.Mock
}

func (m *mockPinService) CreatePin(ctx context.Context, userID uuid.UUID, pin CreatePinRequest) (*UploadImgPinResponse, error) {
	args := m.Called(ctx, userID, pin)
	resp, _ := args.Get(0).(*UploadImgPinResponse)
	return resp, args.Error(1)
}

func (m *mockPinService) GetPinByID(ctx context.Context, id string) (*DownloadImgPinResponse, error) {
	args := m.Called(ctx, id)
	resp, _ := args.Get(0).(*DownloadImgPinResponse)
	return resp, args.Error(1)
}

func (m *mockPinService) GetPins(ctx context.Context, filters PinFilters) ([]DownloadImgPinResponse, int, error) {
	args := m.Called(ctx, filters)
	pins, _ := args.Get(0).([]DownloadImgPinResponse)
	return pins, args.Int(1), args.Error(2)
}

func (m *mockPinService) GetPinsByIDs(ctx context.Context, ids uuid.UUIDs) ([]DownloadImgPinResponse, error) {
	args := m.Called(ctx, ids)
	pins, _ := args.Get(0).([]DownloadImgPinResponse)
	return pins, args.Error(1)
}

func (m *mockPinService) UpdatePin(ctx context.Context, id string, userID uuid.UUID, pin UpdatePinRequest) error {
	args := m.Called(ctx, id, userID, pin)
	return args.Error(0)
}

func (m *mockPinService) DeletePin(ctx context.Context, id string, userID uuid.UUID) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *mockPinService) ConfirmUpload(ctx context.Context, id string, userID uuid.UUID) (*PinResponse, error) {
	args := m.Called(ctx, id, userID)
	resp, _ := args.Get(0).(*PinResponse)
	return resp, args.Error(1)
}
