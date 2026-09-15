package likes

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
)

type LikeService interface {
	AddLike(ctx context.Context, userID, pinID uuid.UUID) error
	RemoveLike(ctx context.Context, userID, pinID uuid.UUID) error
	GetLikeCount(ctx context.Context, pinID uuid.UUID) (*int64, error)
	HasLiked(ctx context.Context, userID, pinID uuid.UUID) (*bool, error)
	HasLikedBatch(ctx context.Context, userID uuid.UUID, pinIDs uuid.UUIDs) (map[uuid.UUID]bool, error)
	ListLikedByUser(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
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
		if err := s.redis.SAdd(ctx, "likes:dirty", pinID.String()).Err(); err != nil {
			return fmt.Errorf("mark pin dirty in redis: %w", err)
		}
		if err := s.redis.Incr(ctx, fmt.Sprintf("likes:count:%s", pinID.String())).Err(); err != nil {
			return fmt.Errorf("incr like count in redis: %w", err)
		}
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

func (s *likeService) GetLikeCount(ctx context.Context, pinID uuid.UUID) (*int64, error) {
	var dbCount *int64
	var redisCount int64

	var g errgroup.Group

	g.Go(func() error {
		c, err := s.repo.GetLikeCount(ctx, pinID)
		if err != nil {
			return fmt.Errorf("get like count from repository: %w", err)
		}
		dbCount = c
		return nil
	})

	g.Go(func() error {
		redisCountStr, err := s.redis.Get(ctx, fmt.Sprintf("likes:count:%s", pinID)).Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				redisCount = 0
				return nil
			}
			return fmt.Errorf("get like count from redis key: %s: %w", pinID, err)
		}
		redisCount, err = strconv.ParseInt(redisCountStr, 10, 64)
		if err != nil {
			return fmt.Errorf("convert redis count %q to int64: %w", redisCountStr, err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	resp := *dbCount + redisCount
	return &resp, nil
}

func (s *likeService) HasLiked(ctx context.Context, userID, pinID uuid.UUID) (*bool, error) {
	hasLiked, err := s.repo.HasLiked(ctx, userID, pinID)
	if err != nil {
		return nil, fmt.Errorf("check like in repository: %w", err)
	}
	return hasLiked, nil
}

func (s *likeService) HasLikedBatch(ctx context.Context, userID uuid.UUID, pinIDs uuid.UUIDs) (map[uuid.UUID]bool, error) {
	lb, err := s.repo.HasLikedBatch(ctx, userID, pinIDs)
	if err != nil {
		return nil, fmt.Errorf("check like batch in repository: %w", err)
	}
	return lb, nil
}

func (s *likeService) ListLikedByUser(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	l, err := s.repo.ListLikedByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("check liked by user in repository: %w", err)
	}
	return l, nil
}
