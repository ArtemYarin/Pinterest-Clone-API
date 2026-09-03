package likes

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LikeRepository interface {
	AddLike(ctx context.Context, userID, pinID uuid.UUID) error
	RemoveLike(ctx context.Context, userID, pinID uuid.UUID) error
}

type likeRepository struct {
	db *pgxpool.Pool
}

func NewLikeRepository(db *pgxpool.Pool) LikeRepository {
	return &likeRepository{db: db}
}

func (r *likeRepository) AddLike(ctx context.Context, userID, pinID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO pin_likes (user_id, pin_id) VALUES ($1, $2)
		 ON CONFLICT (user_id, pin_id) DO NOTHING`,
		userID, pinID)
	if err != nil {
		return fmt.Errorf("AddLike user %v pin %v: %w", userID, pinID, err)
	}
	return nil
}

func (r *likeRepository) RemoveLike(ctx context.Context, userID, pinID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM pin_likes WHERE user_id = $1 AND pin_id = $2`,
		userID, pinID)
	if err != nil {
		return fmt.Errorf("RemoveLike user %v pin %v: %w", userID, pinID, err)
	}
	return nil
}
