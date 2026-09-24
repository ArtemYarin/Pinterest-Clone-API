package profile

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProfileRepository interface {
	CreateProfile(ctx context.Context, userID uuid.UUID, username string) (*Profile, error)
	GetProfileByUserID(ctx context.Context, userID uuid.UUID) (*Profile, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, username, bio *string) error
	SetAvatar(ctx context.Context, userID uuid.UUID, avatarKey, status string) error
}

type profileRepository struct {
	db *pgxpool.Pool
}

func NewProfileRepository(db *pgxpool.Pool) ProfileRepository {
	return &profileRepository{db: db}
}

func (r *profileRepository) CreateProfile(ctx context.Context, userID uuid.UUID, username string) (*Profile, error) {
	var p Profile
	err := r.db.QueryRow(ctx,
		`INSERT INTO profiles (user_id, username) VALUES ($1, $2)
		 ON CONFLICT (user_id) DO NOTHING
		 RETURNING user_id, username, bio, avatar_key, avatar_status, created_at, updated_at`,
		userID, username).
		Scan(&p.UserID, &p.Username, &p.Bio, &p.AvatarKey, &p.AvatarStatus, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if isNotFoundErr(err) {
			// ON CONFLICT DO NOTHING returned no row: profile already exists
			// (e.g. a redelivered registration event) - treat as idempotent success.
			return r.GetProfileByUserID(ctx, userID)
		}
		return nil, fmt.Errorf("CreateProfile user %v: %v", userID, err)
	}
	return &p, nil
}

func (r *profileRepository) GetProfileByUserID(ctx context.Context, userID uuid.UUID) (*Profile, error) {
	var p Profile
	err := r.db.QueryRow(ctx,
		"SELECT user_id, username, bio, avatar_key, avatar_status, created_at, updated_at FROM profiles WHERE user_id = $1",
		userID).
		Scan(&p.UserID, &p.Username, &p.Bio, &p.AvatarKey, &p.AvatarStatus, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if isNotFoundErr(err) {
			return nil, fmt.Errorf("user_id %s not found: %w", userID, errProfileNotFound)
		}
		return nil, fmt.Errorf("GetProfileByUserID user %v: %v", userID, err)
	}
	return &p, nil
}

func (r *profileRepository) UpdateProfile(ctx context.Context, userID uuid.UUID, username, bio *string) error {
	query := "UPDATE profiles SET updated_at = NOW()"
	args := []interface{}{}
	argIndex := 1

	if username != nil {
		query += fmt.Sprintf(", username = $%d", argIndex)
		args = append(args, *username)
		argIndex++
	}
	if bio != nil {
		query += fmt.Sprintf(", bio = $%d", argIndex)
		args = append(args, *bio)
		argIndex++
	}
	if len(args) == 0 {
		return fmt.Errorf("no fields to update: %w", errBadRequest)
	}

	query += fmt.Sprintf(" WHERE user_id = $%d", argIndex)
	args = append(args, userID)

	tag, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("UpdateProfile: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("user_id %s not found: %w", userID, errProfileNotFound)
	}

	return nil
}

func (r *profileRepository) SetAvatar(ctx context.Context, userID uuid.UUID, avatarKey, status string) error {
	tag, err := r.db.Exec(ctx,
		"UPDATE profiles SET avatar_key = $1, avatar_status = $2, updated_at = NOW() WHERE user_id = $3",
		avatarKey, status, userID)
	if err != nil {
		return fmt.Errorf("SetAvatar: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("user_id %s not found: %w", userID, errProfileNotFound)
	}
	return nil
}
