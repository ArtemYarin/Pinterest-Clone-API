package likes

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newTestService(t *testing.T) (*likeService, *mockLikeRepository, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { client.Close() })

	repo := new(mockLikeRepository)
	return &likeService{repo: repo, redis: client}, repo, mr
}

func TestLikeService_AddLike(t *testing.T) {
	userID := uuid.New()
	pinID := uuid.New()

	t.Run("repository error", func(t *testing.T) {
		svc, repo, mr := newTestService(t)
		repoErr := errors.New("boom")
		repo.On("AddLike", mock.Anything, userID, pinID).Return(false, repoErr)

		err := svc.AddLike(context.Background(), userID, pinID)
		require.Error(t, err)
		assert.ErrorIs(t, err, repoErr)
		assert.False(t, mr.Exists("likes:dirty"))
		repo.AssertExpectations(t)
	})

	t.Run("no-op when not inserted", func(t *testing.T) {
		svc, repo, mr := newTestService(t)
		repo.On("AddLike", mock.Anything, userID, pinID).Return(false, nil)

		err := svc.AddLike(context.Background(), userID, pinID)
		require.NoError(t, err)
		assert.False(t, mr.Exists("likes:dirty"))
		repo.AssertExpectations(t)
	})

	t.Run("success updates redis", func(t *testing.T) {
		svc, repo, mr := newTestService(t)
		repo.On("AddLike", mock.Anything, userID, pinID).Return(true, nil)

		err := svc.AddLike(context.Background(), userID, pinID)
		require.NoError(t, err)

		isMember, err := mr.SIsMember("likes:dirty", pinID.String())
		require.NoError(t, err)
		assert.True(t, isMember)

		countKey := fmt.Sprintf("likes:count:%s", pinID.String())
		got, err := mr.Get(countKey)
		require.NoError(t, err)
		assert.Equal(t, "1", got)
		repo.AssertExpectations(t)
	})
}

func TestLikeService_RemoveLike(t *testing.T) {
	userID := uuid.New()
	pinID := uuid.New()

	t.Run("repository error", func(t *testing.T) {
		svc, repo, mr := newTestService(t)
		repoErr := errors.New("boom")
		repo.On("RemoveLike", mock.Anything, userID, pinID).Return(false, repoErr)

		err := svc.RemoveLike(context.Background(), userID, pinID)
		require.Error(t, err)
		assert.ErrorIs(t, err, repoErr)
		assert.False(t, mr.Exists("likes:dirty"))
		repo.AssertExpectations(t)
	})

	t.Run("no-op when not deleted", func(t *testing.T) {
		svc, repo, mr := newTestService(t)
		repo.On("RemoveLike", mock.Anything, userID, pinID).Return(false, nil)

		err := svc.RemoveLike(context.Background(), userID, pinID)
		require.NoError(t, err)
		assert.False(t, mr.Exists("likes:dirty"))
		repo.AssertExpectations(t)
	})

	t.Run("success updates redis", func(t *testing.T) {
		svc, repo, mr := newTestService(t)
		repo.On("RemoveLike", mock.Anything, userID, pinID).Return(true, nil)

		err := svc.RemoveLike(context.Background(), userID, pinID)
		require.NoError(t, err)

		isMember, err := mr.SIsMember("likes:dirty", pinID.String())
		require.NoError(t, err)
		assert.True(t, isMember)

		countKey := fmt.Sprintf("likes:count:%s", pinID.String())
		got, err := mr.Get(countKey)
		require.NoError(t, err)
		assert.Equal(t, "-1", got)
		repo.AssertExpectations(t)
	})
}

func TestLikeService_GetLikeCount(t *testing.T) {
	pinID := uuid.New()

	t.Run("repository error", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		repoErr := errors.New("boom")
		repo.On("GetLikeCount", mock.Anything, pinID).Return((*int64)(nil), repoErr)

		got, err := svc.GetLikeCount(context.Background(), pinID)
		require.Error(t, err)
		assert.ErrorIs(t, err, repoErr)
		assert.Nil(t, got)
		repo.AssertExpectations(t)
	})

	t.Run("sums db count and redis delta", func(t *testing.T) {
		svc, repo, mr := newTestService(t)
		dbCount := int64(5)
		repo.On("GetLikeCount", mock.Anything, pinID).Return(&dbCount, nil)
		mr.Set(fmt.Sprintf("likes:count:%s", pinID), "3")

		got, err := svc.GetLikeCount(context.Background(), pinID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, int64(8), *got)
		repo.AssertExpectations(t)
	})

	t.Run("missing redis key contributes zero", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		dbCount := int64(5)
		repo.On("GetLikeCount", mock.Anything, pinID).Return(&dbCount, nil)

		got, err := svc.GetLikeCount(context.Background(), pinID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, int64(5), *got)
		repo.AssertExpectations(t)
	})
}

func TestLikeService_HasLiked(t *testing.T) {
	userID := uuid.New()
	pinID := uuid.New()

	t.Run("repository error", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		repoErr := errors.New("boom")
		repo.On("HasLiked", mock.Anything, userID, pinID).Return((*bool)(nil), repoErr)

		got, err := svc.HasLiked(context.Background(), userID, pinID)
		require.Error(t, err)
		assert.ErrorIs(t, err, repoErr)
		assert.Nil(t, got)
		repo.AssertExpectations(t)
	})

	t.Run("success passes through result", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		liked := true
		repo.On("HasLiked", mock.Anything, userID, pinID).Return(&liked, nil)

		got, err := svc.HasLiked(context.Background(), userID, pinID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.True(t, *got)
		repo.AssertExpectations(t)
	})
}

func TestLikeService_HasLikedBatch(t *testing.T) {
	userID := uuid.New()
	pinIDs := uuid.UUIDs{uuid.New(), uuid.New()}

	t.Run("repository error", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		repoErr := errors.New("boom")
		repo.On("HasLikedBatch", mock.Anything, userID, pinIDs).Return(map[uuid.UUID]bool(nil), repoErr)

		got, err := svc.HasLikedBatch(context.Background(), userID, pinIDs)
		require.Error(t, err)
		assert.ErrorIs(t, err, repoErr)
		assert.Nil(t, got)
		repo.AssertExpectations(t)
	})

	t.Run("success passes through result", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		expected := map[uuid.UUID]bool{pinIDs[0]: true, pinIDs[1]: false}
		repo.On("HasLikedBatch", mock.Anything, userID, pinIDs).Return(expected, nil)

		got, err := svc.HasLikedBatch(context.Background(), userID, pinIDs)
		require.NoError(t, err)
		assert.Equal(t, expected, got)
		repo.AssertExpectations(t)
	})
}

func TestLikeService_ListLikedByUser(t *testing.T) {
	userID := uuid.New()

	t.Run("repository error", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		repoErr := errors.New("boom")
		repo.On("ListLikedByUser", mock.Anything, userID).Return([]uuid.UUID(nil), repoErr)

		got, err := svc.ListLikedByUser(context.Background(), userID)
		require.Error(t, err)
		assert.ErrorIs(t, err, repoErr)
		assert.Nil(t, got)
		repo.AssertExpectations(t)
	})

	t.Run("success passes through result", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		expected := []uuid.UUID{uuid.New(), uuid.New()}
		repo.On("ListLikedByUser", mock.Anything, userID).Return(expected, nil)

		got, err := svc.ListLikedByUser(context.Background(), userID)
		require.NoError(t, err)
		assert.Equal(t, expected, got)
		repo.AssertExpectations(t)
	})
}
