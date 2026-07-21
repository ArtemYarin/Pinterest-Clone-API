package auth

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user CredentialsUserRequest) (*UserResponse, error)
	GetUserByEmail(ctx context.Context, email string) (*UserWithPasswordResponse, error)
	GetUserByID(ctx context.Context, id string) (*UserResponse, error)
	UpdateUser(ctx context.Context, user UpdateUserRequest) error
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
		return fmt.Errorf("no fields to update: %w", errValidation)
	}

	query += fmt.Sprintf(" WHERE id = $%d", argIndex)
	args = append(args, user.Id)

	_, err := r.db.Exec(ctx, query, args...)

	return fmt.Errorf("UpdateUser: %w", err)
}
