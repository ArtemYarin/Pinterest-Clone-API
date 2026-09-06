package likes

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type LikeService interface {
	AddLike(ctx context.Context, userID, pinID uuid.UUID) error
	RemoveLike(ctx context.Context, userID, pinID uuid.UUID) error
}

type likeService struct {
	repo  LikeRepository
	redis *redis.Client
}

func NewLikeService(repo LikeRepository, redis *redis.Client) LikeService {
	return &likeService{repo: repo, redis: redis}
}

func (s *likeService) AddLike(ctx context.Context, userID, pinID uuid.UUID) error {
	inserted, err := s.repo.AddLike(ctx, userID, pinID)
	if err != nil {
		return fmt.Errorf("add like in repository: %w", err)
	}
	if inserted {
		s.redis.SAdd(ctx, "likes:dirty", pinID.String())
		s.redis.Incr(ctx, fmt.Sprintf("likes:count:%s", pinID.String()))
	}
	return nil
}

func (s *likeService) RemoveLike(ctx context.Context, userID, pinID uuid.UUID) error {
	deleted, err := s.repo.RemoveLike(ctx, userID, pinID)
	if err != nil {
		return fmt.Errorf("remove like in repository: %w", err)
	}
	if deleted {
		s.redis.SAdd(ctx, "likes:dirty", pinID.String())
		s.redis.Decr(ctx, fmt.Sprintf("likes:count:%s", pinID.String()))
	}
	return nil
}
