package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user CredentialsUserRequest) (*UserResponse, error)
	GetUserByEmail(ctx context.Context, email string) (*UserWithPasswordResponse, error)
	GetUserByID(ctx context.Context, id string) (*UserResponse, error)
	UpdateUser(ctx context.Context, user UpdateUserRequest) error
	SaveRefreshToken(ctx context.Context, user_id, hash string, ttl time.Time) error
	FindRefreshToken(ctx context.Context, hash string) (*RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, id uuid.UUID) error
	RevokeRefreshTokenByHash(ctx context.Context, hash string) error
}

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) CreateUser(ctx context.Context, user CredentialsUserRequest) (*UserResponse, error) {
	var u UserResponse
	err := r.db.QueryRow(ctx,
		"INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id, email, created_at, updated_at",
		user.Email, user.Password_hash).
		Scan(&u.Id, &u.Email, &u.Created_at, &u.Updated_at)
	if err != nil {
		if isDuplicateErr(err) {
			return nil, fmt.Errorf("email %v already exists: %w", user.Email, errEmailExists)
		}
		return nil, fmt.Errorf("CreateUser %v: %v", user, err)
	}
	return &u, nil
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*UserWithPasswordResponse, error) {
	var u UserWithPasswordResponse
	err := r.db.QueryRow(ctx,
		"SELECT id, email, password_hash, created_at, updated_at FROM users WHERE email = $1", email).
		Scan(&u.Id, &u.Email, &u.Password_hash, &u.Created_at, &u.Updated_at)
	if err != nil {
		if isNotFoundErr(err) {
			return nil, fmt.Errorf("email %s not found: %w", email, errUserNotFound)
		}
		return nil, fmt.Errorf("GetUserByEmail email %v: %v", email, err)
	}
	return &u, nil
}

func (r *userRepository) GetUserByID(ctx context.Context, id string) (*UserResponse, error) {
	var u UserResponse
	err := r.db.QueryRow(ctx,
		"SELECT id, email, created_at, updated_at FROM users WHERE id = $1", id).
		Scan(&u.Id, &u.Email, &u.Created_at, &u.Updated_at)
	if err != nil {
		if isNotFoundErr(err) {
			return nil, fmt.Errorf("id %s not found: %w", id, errUserNotFound)
		}
		return nil, fmt.Errorf("GetUserByID id %v: %v", id, err)
	}
	return &u, nil
}

func (r *userRepository) UpdateUser(ctx context.Context, user UpdateUserRequest) error {
	query := "UPDATE users SET updated_at = NOW()"
	args := []interface{}{}
	argIndex := 1

	if user.Email != nil {
		query += fmt.Sprintf(", email = $%d", argIndex)
		args = append(args, *user.Email)
		argIndex++
	}
	if user.Password_hash != nil {
		query += fmt.Sprintf(", password_hash = $%d", argIndex)
		args = append(args, *user.Password_hash)
		argIndex++
	}
	if len(args) == 0 {
		return fmt.Errorf("no fields to update: %w", errBadRequest)
	}

	query += fmt.Sprintf(" WHERE id = $%d", argIndex)
	args = append(args, user.Id)

	_, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("UpdateUser: %w", err)
	}

	return nil
}

func (r *userRepository) SaveRefreshToken(ctx context.Context, user_id, hash string, ttl time.Time) error {
	_, err := r.db.Exec(ctx,
		"INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)",
		user_id, hash, ttl)

	if err != nil {
		return fmt.Errorf("SaveRefreshToken user_id %s: %v", user_id, err)
	}

	return nil
}

func (r *userRepository) FindRefreshToken(ctx context.Context, hash string) (*RefreshToken, error) {
	var t RefreshToken
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, token_hash, expires_at, revoked_at, created_at 
		FROM refresh_tokens
		WHERE token_hash = $1`, hash,
	).Scan(&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &t.RevokedAt, &t.CreatedAt)

	if err != nil {
		if isNotFoundErr(err) {
			return nil, fmt.Errorf("refresh token not found: %w", errTokenNotFound)
		}
		return nil, fmt.Errorf("FindRefreshToken: %v", err)
	}

	return &t, nil
}

// RevokeRefreshToken marks a token as revoked. Only an unrevoked token matches,
// so of two concurrent refreshes with the same token only one succeeds.
func (r *userRepository) RevokeRefreshToken(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx,
		"UPDATE refresh_tokens SET revoked_at = now() WHERE id = $1 AND revoked_at IS NULL", id)
	if err != nil {
		return fmt.Errorf("RevokeRefreshToken id %s: %v", id, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("refresh token %s not found or already revoked: %w", id, errTokenNotFound)
	}
	return nil
}

func (r *userRepository) RevokeRefreshTokenByHash(ctx context.Context, hash string) error {
	tag, err := r.db.Exec(ctx,
		"UPDATE refresh_tokens SET revoked_at = now() WHERE token_hash = $1 AND revoked_at IS NULL", hash)
	if err != nil {
		return fmt.Errorf("RevokeRefreshTokenByHash: %v", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("refresh token not found or already revoked: %w", errTokenNotFound)
	}
	return nil
}

type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time // nil = still valid
	CreatedAt time.Time
}

func (t *RefreshToken) IsValid(now time.Time) bool {
	return t.RevokedAt == nil && now.Before(t.ExpiresAt)
}
