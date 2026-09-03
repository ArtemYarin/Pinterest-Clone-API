package likes

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type LikeService interface {
	AddLike(ctx context.Context, userID, pinID uuid.UUID) error
	RemoveLike(ctx context.Context, userID, pinID uuid.UUID) error
}

type likeService struct {
	repo LikeRepository
}

func NewLikeService(repo LikeRepository) LikeService {
	return &likeService{repo: repo}
}

func (s *likeService) AddLike(ctx context.Context, userID, pinID uuid.UUID) error {
	if err := s.repo.AddLike(ctx, userID, pinID); err != nil {
		return fmt.Errorf("add like in repository: %w", err)
	}
	return nil
}

func (s *likeService) RemoveLike(ctx context.Context, userID, pinID uuid.UUID) error {
	if err := s.repo.RemoveLike(ctx, userID, pinID); err != nil {
		return fmt.Errorf("remove like in repository: %w", err)
	}
	return nil
}
