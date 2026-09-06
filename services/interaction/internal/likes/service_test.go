package likes

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type mockLikeRepository struct {
	addLikeFunc    func(ctx context.Context, userID, pinID uuid.UUID) (bool, error)
	removeLikeFunc func(ctx context.Context, userID, pinID uuid.UUID) (bool, error)
}

func (m *mockLikeRepository) AddLike(ctx context.Context, userID, pinID uuid.UUID) (bool, error) {
	return m.addLikeFunc(ctx, userID, pinID)
}

func (m *mockLikeRepository) RemoveLike(ctx context.Context, userID, pinID uuid.UUID) (bool, error) {
	return m.removeLikeFunc(ctx, userID, pinID)
}

func newTestService(t *testing.T, repo LikeRepository) (*likeService, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { client.Close() })
	return &likeService{repo: repo, redis: client}, mr
}

func TestLikeService_AddLike(t *testing.T) {
	userID := uuid.New()
	pinID := uuid.New()

	t.Run("repository error", func(t *testing.T) {
		repoErr := errors.New("boom")
		svc, mr := newTestService(t, &mockLikeRepository{
			addLikeFunc: func(ctx context.Context, userID, pinID uuid.UUID) (bool, error) {
				return false, repoErr
			},
		})

		err := svc.AddLike(context.Background(), userID, pinID)
		if err == nil || !errors.Is(err, repoErr) {
			t.Errorf("expected wrapped repo error, got %v", err)
		}
		if mr.Exists("likes:dirty") {
			t.Errorf("expected redis untouched on repository error")
		}
	})

	t.Run("no-op when not inserted", func(t *testing.T) {
		svc, mr := newTestService(t, &mockLikeRepository{
			addLikeFunc: func(ctx context.Context, userID, pinID uuid.UUID) (bool, error) {
				return false, nil
			},
		})

		if err := svc.AddLike(context.Background(), userID, pinID); err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
		if mr.Exists("likes:dirty") {
			t.Errorf("expected redis untouched when nothing was inserted")
		}
	})

	t.Run("success updates redis", func(t *testing.T) {
		svc, mr := newTestService(t, &mockLikeRepository{
			addLikeFunc: func(ctx context.Context, userID, pinID uuid.UUID) (bool, error) {
				return true, nil
			},
		})

		if err := svc.AddLike(context.Background(), userID, pinID); err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
		if isMember, err := mr.SIsMember("likes:dirty", pinID.String()); err != nil || !isMember {
			t.Errorf("expected pinID to be added to likes:dirty set, err=%v", err)
		}
		countKey := fmt.Sprintf("likes:count:%s", pinID.String())
		if got, err := mr.Get(countKey); err != nil || got != "1" {
			t.Errorf("expected %s to be 1, got %q, err=%v", countKey, got, err)
		}
	})
}

func TestLikeService_RemoveLike(t *testing.T) {
	userID := uuid.New()
	pinID := uuid.New()

	t.Run("repository error", func(t *testing.T) {
		repoErr := errors.New("boom")
		svc, mr := newTestService(t, &mockLikeRepository{
			removeLikeFunc: func(ctx context.Context, userID, pinID uuid.UUID) (bool, error) {
				return false, repoErr
			},
		})

		err := svc.RemoveLike(context.Background(), userID, pinID)
		if err == nil || !errors.Is(err, repoErr) {
			t.Errorf("expected wrapped repo error, got %v", err)
		}
		if mr.Exists("likes:dirty") {
			t.Errorf("expected redis untouched on repository error")
		}
	})

	t.Run("no-op when not deleted", func(t *testing.T) {
		svc, mr := newTestService(t, &mockLikeRepository{
			removeLikeFunc: func(ctx context.Context, userID, pinID uuid.UUID) (bool, error) {
				return false, nil
			},
		})

		if err := svc.RemoveLike(context.Background(), userID, pinID); err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
		if mr.Exists("likes:dirty") {
			t.Errorf("expected redis untouched when nothing was deleted")
		}
	})

	t.Run("success updates redis", func(t *testing.T) {
		svc, mr := newTestService(t, &mockLikeRepository{
			removeLikeFunc: func(ctx context.Context, userID, pinID uuid.UUID) (bool, error) {
				return true, nil
			},
		})

		if err := svc.RemoveLike(context.Background(), userID, pinID); err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
		if isMember, err := mr.SIsMember("likes:dirty", pinID.String()); err != nil || !isMember {
			t.Errorf("expected pinID to be added to likes:dirty set, err=%v", err)
		}
		countKey := fmt.Sprintf("likes:count:%s", pinID.String())
		if got, err := mr.Get(countKey); err != nil || got != "-1" {
			t.Errorf("expected %s to be -1, got %q, err=%v", countKey, got, err)
		}
	})
}
