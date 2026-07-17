package auth

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user CredentialsUserRequest) (*UserCreatedResponse, error)
	GetUserByEmail(ctx context.Context, email string) (*UserDataResponse, error)
	GetUserByID(ctx context.Context, id string) (*UserDataResponse, error)
	UpdateUser(ctx context.Context, user UpdateUserRequest) error
}

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) CreateUser(ctx context.Context, user CredentialsUserRequest) (*UserCreatedResponse, error) {
	var u UserCreatedResponse
	err := r.db.QueryRow(ctx,
		"INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id, email, created_at, updated_at",
		user.Email, user.Password_hash).
		Scan(&u.Id, &u.Email, &u.Created_at, &u.Updated_at)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*UserDataResponse, error) {
	var u UserDataResponse
	err := r.db.QueryRow(ctx,
		"SELECT id, email, password_hash, created_at, updated_at FROM users WHERE email = $1", email).
		Scan(&u.Id, &u.Email, &u.Password_hash, &u.Created_at, &u.Updated_at)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) GetUserByID(ctx context.Context, id string) (*UserDataResponse, error) {
	var u UserDataResponse
	err := r.db.QueryRow(ctx,
		"SELECT id, email, password_hash, created_at, updated_at FROM users WHERE id = $1", id).
		Scan(&u.Id, &u.Email, &u.Password_hash, &u.Created_at, &u.Updated_at)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) UpdateUser(ctx context.Context, user UpdateUserRequest) error {
	query := "UPDATE users SET updated_at = NOW()"
	if user.Email != nil {
		query += fmt.Sprintf(", SET email = %s", *user.Email)
	}
	if user.Password_hash != nil {
		query += fmt.Sprintf(", SET password_hash = %s", *user.Password_hash)
	}
	query += fmt.Sprintf(" WHERE id = %s", user.Id)

	_, err := r.db.Exec(ctx, query)

	return err
}
