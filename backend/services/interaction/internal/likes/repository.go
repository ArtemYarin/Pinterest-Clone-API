package likes

import (
	"context"
	"fmt"

	"github.com/ArtemYarin/pinterest-clone-api/services/interaction-service/internal/shared/errs"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LikeRepository interface {
	AddLike(ctx context.Context, userID, pinID uuid.UUID) (bool, error)
	RemoveLike(ctx context.Context, userID, pinID uuid.UUID) (bool, error)
	GetLikeCount(ctx context.Context, pinID uuid.UUID) (*int64, error)
	HasLiked(ctx context.Context, userID, pinID uuid.UUID) (*bool, error)
	HasLikedBatch(ctx context.Context, userID uuid.UUID, pinIDs uuid.UUIDs) (map[uuid.UUID]bool, error)
	ListLikedByUser(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
}

type likeRepository struct {
	db *pgxpool.Pool
}

func NewLikeRepository(db *pgxpool.Pool) LikeRepository {
	return &likeRepository{db: db}
}

// AddLike adds creates pin_likes instance if it didn't exist before.
// It returns true if pin_likes was created and false if didn't (for example: conflict)
func (r *likeRepository) AddLike(ctx context.Context, userID, pinID uuid.UUID) (bool, error) {
	tag, err := r.db.Exec(ctx,
		`INSERT INTO pin_likes (user_id, pin_id) VALUES ($1, $2)
		 ON CONFLICT (user_id, pin_id) DO NOTHING`,
		userID, pinID)
	if err != nil {
		return false, fmt.Errorf("AddLike user %v pin %v: %w", userID, pinID, err)
	}

	inserted := false

	if tag.RowsAffected() != 0 {
		inserted = true
	}
	return inserted, nil
}

// RemoveLike deletes pin_likes instance if it exists.
// It returns true if pin_likes was deleted and false if didn't
func (r *likeRepository) RemoveLike(ctx context.Context, userID, pinID uuid.UUID) (bool, error) {
	tag, err := r.db.Exec(ctx,
		`DELETE FROM pin_likes WHERE user_id = $1 AND pin_id = $2`,
		userID, pinID)
	if err != nil {
		return false, fmt.Errorf("RemoveLike user %v pin %v: %w", userID, pinID, err)
	}

	deleted := false
	if tag.RowsAffected() != 0 {
		deleted = true
	}
	return deleted, nil
}

// GetLikeCount selects count row from like_counts.
// It returns 0 if like_counts with given pinID doesn't exist
func (r *likeRepository) GetLikeCount(ctx context.Context, pinID uuid.UUID) (*int64, error) {
	var count int64
	err := r.db.QueryRow(ctx,
		"SELECT count FROM like_counts WHERE target_id = $1", pinID).
		Scan(&count)
	if err != nil {
		if errs.IsNotFoundErr(err) {
			zero := int64(0)
			return &zero, nil
		}
		return nil, fmt.Errorf("GetLikeCount pin ID: %v: %w", pinID, err)
	}
	return &count, nil
}

// HasLiked check pin_likes existens for given userID and pinID
func (r *likeRepository) HasLiked(ctx context.Context, userID, pinID uuid.UUID) (*bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		"SELECT EXISTS (SELECT  1 FROM pin_likes WHERE user_id = $1 AND pin_id = $2) AS like_exists",
		userID, pinID).
		Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("HasLiked user ID %v pin ID: %v: %w", userID, pinID, err)
	}
	return &exists, nil
}

// HasLikedBatch check pin_likes existens for given userID and pinIDs.
// Retuens a map where key is pinID and value is bool (true - liked, false - didn't)
func (r *likeRepository) HasLikedBatch(ctx context.Context, userID uuid.UUID, pinIDs uuid.UUIDs) (map[uuid.UUID]bool, error) {
	result := make(map[uuid.UUID]bool, len(pinIDs))
	if len(pinIDs) == 0 {
		return result, nil
	}
	// Initialize all as false by default
	for _, id := range pinIDs {
		result[id] = false
	}

	rows, err := r.db.Query(ctx,
		"SELECT pin_id FROM pin_likes WHERE pin_id = ANY($1) AND user_id = $2",
		pinIDs, userID)
	if err != nil {
		return nil, fmt.Errorf("HasLikedBatch user ID %v (%d pin IDs): %w", userID, len(pinIDs), err)
	}
	defer rows.Close()

	for rows.Next() {
		var pinID uuid.UUID
		if err := rows.Scan(&pinID); err != nil {
			return nil, fmt.Errorf("HasLikedBatch: scan liked row: %w", err)
		}
		result[pinID] = true
	}
	return result, rows.Err()
}

// ListLikedByUser returns a slice with all pinIDs that user liked
func (r *likeRepository) ListLikedByUser(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.db.Query(ctx,
		"SELECT pin_id FROM pin_likes WHERE user_id = $1",
		userID)

	if err != nil {
		return nil, fmt.Errorf("ListLikedByUser user ID %v: %w", userID, err)
	}
	defer rows.Close()

	var result []uuid.UUID
	for rows.Next() {
		var pinID uuid.UUID
		if err := rows.Scan(&pinID); err != nil {
			return nil, fmt.Errorf("ListLikedByUser: scan liked row: %w", err)
		}
		result = append(result, pinID)
	}
	return result, rows.Err()
}
