package likes

import (
	"context"
	"fmt"

	"github.com/google/uuid"
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
	repo LikeRepository
}

func NewLikeService(repo LikeRepository) LikeService {
	return &likeService{repo: repo}
}

func (s *likeService) AddLike(ctx context.Context, userID, pinID uuid.UUID) error {
	if _, err := s.repo.AddLike(ctx, userID, pinID); err != nil {
		return fmt.Errorf("add like in repository: %w", err)
	}
	return nil
}

func (s *likeService) RemoveLike(ctx context.Context, userID, pinID uuid.UUID) error {
	if _, err := s.repo.RemoveLike(ctx, userID, pinID); err != nil {
		return fmt.Errorf("remove like in repository: %w", err)
	}
	return nil
}

func (s *likeService) GetLikeCount(ctx context.Context, pinID uuid.UUID) (*int64, error) {
	c, err := s.repo.GetLikeCount(ctx, pinID)
	if err != nil {
		return nil, fmt.Errorf("get like count from repository: %w", err)
	}
	return c, nil
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
