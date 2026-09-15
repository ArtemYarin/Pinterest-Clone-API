package likes

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type mockLikeRepository struct {
	mock.Mock
}

func (m *mockLikeRepository) AddLike(ctx context.Context, userID, pinID uuid.UUID) (bool, error) {
	args := m.Called(ctx, userID, pinID)
	return args.Bool(0), args.Error(1)
}

func (m *mockLikeRepository) RemoveLike(ctx context.Context, userID, pinID uuid.UUID) (bool, error) {
	args := m.Called(ctx, userID, pinID)
	return args.Bool(0), args.Error(1)
}

func (m *mockLikeRepository) GetLikeCount(ctx context.Context, pinID uuid.UUID) (*int64, error) {
	args := m.Called(ctx, pinID)
	count, _ := args.Get(0).(*int64)
	return count, args.Error(1)
}

func (m *mockLikeRepository) HasLiked(ctx context.Context, userID, pinID uuid.UUID) (*bool, error) {
	args := m.Called(ctx, userID, pinID)
	liked, _ := args.Get(0).(*bool)
	return liked, args.Error(1)
}

func (m *mockLikeRepository) HasLikedBatch(ctx context.Context, userID uuid.UUID, pinIDs uuid.UUIDs) (map[uuid.UUID]bool, error) {
	args := m.Called(ctx, userID, pinIDs)
	result, _ := args.Get(0).(map[uuid.UUID]bool)
	return result, args.Error(1)
}

func (m *mockLikeRepository) ListLikedByUser(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	args := m.Called(ctx, userID)
	result, _ := args.Get(0).([]uuid.UUID)
	return result, args.Error(1)
}

type mockLikeService struct {
	mock.Mock
}

func (m *mockLikeService) AddLike(ctx context.Context, userID, pinID uuid.UUID) error {
	args := m.Called(ctx, userID, pinID)
	return args.Error(0)
}

func (m *mockLikeService) RemoveLike(ctx context.Context, userID, pinID uuid.UUID) error {
	args := m.Called(ctx, userID, pinID)
	return args.Error(0)
}

func (m *mockLikeService) GetLikeCount(ctx context.Context, pinID uuid.UUID) (*int64, error) {
	args := m.Called(ctx, pinID)
	count, _ := args.Get(0).(*int64)
	return count, args.Error(1)
}

func (m *mockLikeService) HasLiked(ctx context.Context, userID, pinID uuid.UUID) (*bool, error) {
	args := m.Called(ctx, userID, pinID)
	liked, _ := args.Get(0).(*bool)
	return liked, args.Error(1)
}

func (m *mockLikeService) HasLikedBatch(ctx context.Context, userID uuid.UUID, pinIDs uuid.UUIDs) (map[uuid.UUID]bool, error) {
	args := m.Called(ctx, userID, pinIDs)
	result, _ := args.Get(0).(map[uuid.UUID]bool)
	return result, args.Error(1)
}

func (m *mockLikeService) ListLikedByUser(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	args := m.Called(ctx, userID)
	result, _ := args.Get(0).([]uuid.UUID)
	return result, args.Error(1)
}
