package pin

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type PinService interface {
	CreatePin(ctx context.Context, userID uuid.UUID, pin CreatePinRequest) (*PinResponse, error)
	GetPinByID(ctx context.Context, id string) (*PinResponse, error)
	GetPins(ctx context.Context, filters PinFilters) ([]*PinResponse, int, error)
	UpdatePin(ctx context.Context, id string, userID uuid.UUID, pin UpdatePinRequest) error
	DeletePin(ctx context.Context, id string, userID uuid.UUID) error
}

type pinService struct {
	repo     PinRepository
	validate *validator.Validate
}

func NewPinService(repo PinRepository, validate *validator.Validate) PinService {
	return &pinService{repo: repo, validate: validate}
}

func (s *pinService) CreatePin(ctx context.Context, userID uuid.UUID, pin CreatePinRequest) (*PinResponse, error) {
	// Validate input
	err := s.validate.Struct(pin)
	if err != nil {
		valErr := newValidationErr(getValidationMap(err))
		return nil, fmt.Errorf("input validation: %w", valErr)
	}

	// Repository call
	p, err := s.repo.CreatePin(ctx, userID, pin)
	if err != nil {
		return nil, fmt.Errorf("create pin in repository: %w", err)
	}
	return p, nil
}

func (s *pinService) GetPinByID(ctx context.Context, id string) (*PinResponse, error) {
	p, err := s.repo.GetPinByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get pin from repository: %w", err)
	}
	return p, nil
}

func (s *pinService) GetPins(ctx context.Context, filters PinFilters) ([]*PinResponse, int, error) {
	pins, count, err := s.repo.GetPins(ctx, filters)
	if err != nil {
		return nil, 0, fmt.Errorf("get pins from repository: %w", err)
	}
	return pins, count, nil
}

func (s *pinService) UpdatePin(ctx context.Context, id string, userID uuid.UUID, pin UpdatePinRequest) error {
	// Validate input
	err := s.validate.Struct(pin)
	if err != nil {
		valErr := newValidationErr(getValidationMap(err))
		return fmt.Errorf("input validation: %w", valErr)
	}

	// Ownership check
	existing, err := s.repo.GetPinByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get pin from repository: %w", err)
	}
	if existing.User_id != userID {
		return fmt.Errorf("update pin: %w", errForbidden)
	}

	parsedID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("parse id: %w", errBadRequest)
	}
	pin.Id = parsedID

	// Repository call
	err = s.repo.UpdatePin(ctx, pin)
	if err != nil {
		return fmt.Errorf("update pin in repository: %w", err)
	}
	return nil
}

func (s *pinService) DeletePin(ctx context.Context, id string, userID uuid.UUID) error {
	// Ownership check
	existing, err := s.repo.GetPinByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get pin from repository: %w", err)
	}
	if existing.User_id != userID {
		return fmt.Errorf("delete pin: %w", errForbidden)
	}

	// Repository call
	err = s.repo.DeletePin(ctx, id)
	if err != nil {
		return fmt.Errorf("delete pin in repository: %w", err)
	}
	return nil
}
