package pin

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PinRepository interface {
	CreatePin(ctx context.Context, userID uuid.UUID, pin CreatePinRequest) (*PinResponse, error)
	GetPinByID(ctx context.Context, id string) (*PinResponse, error)
	UpdatePin(ctx context.Context, pin UpdatePinRequest) error
	DeletePin(ctx context.Context, id string) error
}

type pinRepository struct {
	db *pgxpool.Pool
}

func NewPinRepository(db *pgxpool.Pool) PinRepository {
	return &pinRepository{db: db}
}

func (r *pinRepository) CreatePin(ctx context.Context, userID uuid.UUID, pin CreatePinRequest) (*PinResponse, error) {
	var p PinResponse
	err := r.db.QueryRow(ctx,
		`INSERT INTO pins (user_id, title, image_url, description)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, user_id, title, image_url, description, created_at, updated_at, likes_count`,
		userID, pin.Title, pin.Image_url, pin.Description).
		Scan(&p.Id, &p.User_id, &p.Title, &p.Image_url, &p.Description, &p.Created_at, &p.Updated_at, &p.Likes_count)
	if err != nil {
		if isDuplicateErr(err) {
			return nil, fmt.Errorf("image url %v already exists: %w", pin.Image_url, errImageURLExists)
		}
		return nil, fmt.Errorf("CreatePin %v: %v", pin, err)
	}
	return &p, nil
}

func (r *pinRepository) GetPinByID(ctx context.Context, id string) (*PinResponse, error) {
	var p PinResponse
	err := r.db.QueryRow(ctx,
		`SELECT id, user_id, title, image_url, description, created_at, updated_at, likes_count
		 FROM pins WHERE id = $1`, id).
		Scan(&p.Id, &p.User_id, &p.Title, &p.Image_url, &p.Description, &p.Created_at, &p.Updated_at, &p.Likes_count)
	if err != nil {
		if isNotFoundErr(err) {
			return nil, fmt.Errorf("id %s not found: %w", id, errPinNotFound)
		}
		return nil, fmt.Errorf("GetPinByID id %v: %v", id, err)
	}
	return &p, nil
}

func (r *pinRepository) UpdatePin(ctx context.Context, pin UpdatePinRequest) error {
	query := "UPDATE pins SET updated_at = NOW()"
	args := []interface{}{}
	argIndex := 1

	if pin.Title != nil {
		query += fmt.Sprintf(", title = $%d", argIndex)
		args = append(args, *pin.Title)
		argIndex++
	}
	if pin.Image_url != nil {
		query += fmt.Sprintf(", image_url = $%d", argIndex)
		args = append(args, *pin.Image_url)
		argIndex++
	}
	if pin.Description != nil {
		query += fmt.Sprintf(", description = $%d", argIndex)
		args = append(args, *pin.Description)
		argIndex++
	}
	if len(args) == 0 {
		return fmt.Errorf("no fields to update: %w", errBadRequest)
	}

	query += fmt.Sprintf(" WHERE id = $%d", argIndex)
	args = append(args, pin.Id)

	tag, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		if isDuplicateErr(err) {
			return fmt.Errorf("image url already exists: %w", errImageURLExists)
		}
		return fmt.Errorf("UpdatePin: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("id %s not found: %w", pin.Id, errPinNotFound)
	}

	return nil
}

func (r *pinRepository) DeletePin(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, "DELETE FROM pins WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("DeletePin: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("id %s not found: %w", id, errPinNotFound)
	}
	return nil
}
