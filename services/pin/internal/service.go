package pin

import (
	"context"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type PinService interface {
	CreatePin(ctx context.Context, userID uuid.UUID, pin CreatePinRequest) (*UploadImgPinResponse, error)
	GetPinByID(ctx context.Context, id string) (*DownloadImgPinResponse, error)
	GetPins(ctx context.Context, filters PinFilters) ([]*PinResponse, int, error)
	UpdatePin(ctx context.Context, id string, userID uuid.UUID, pin UpdatePinRequest) error
	DeletePin(ctx context.Context, id string, userID uuid.UUID) error
}

type pinService struct {
	repo       PinRepository
	validate   *validator.Validate
	imgStorage *ImageStorage
}

func NewPinService(repo PinRepository, validate *validator.Validate, imgStorage *ImageStorage) PinService {
	return &pinService{repo: repo, validate: validate, imgStorage: imgStorage}
}

func (s *pinService) CreatePin(ctx context.Context, userID uuid.UUID, pin CreatePinRequest) (*UploadImgPinResponse, error) {
	// Validate input
	err := s.validate.Struct(pin)
	if err != nil {
		valErr := newValidationErr(getValidationMap(err))
		return nil, fmt.Errorf("input validation: %w", valErr)
	}

	// Upload image
	imageKey := fmt.Sprintf("pins/%s/%s", userID, uuid.New().String())              // MiniO identifier
	uploadURL, err := s.imgStorage.GenerateUploadURL(ctx, imageKey, 15*time.Minute) // temporary upload url
	if err != nil {
		return nil, fmt.Errorf("generate image upload URL: %w", err)
	}

	// Repository call
	p, err := s.repo.CreatePin(ctx, userID, imageKey, pin)
	if err != nil {
		return nil, fmt.Errorf("create pin in repository: %w", err)
	}

	// Response
	resp := UploadImgPinResponse{
		Pin:        *p,
		Upload_url: uploadURL,
	}

	return &resp, nil
}

func (s *pinService) GetPinByID(ctx context.Context, id string) (*DownloadImgPinResponse, error) {
	pin, err := s.repo.GetPinByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get pin from repository: %w", err)
	}

	// Download link
	downloadURL, err := s.imgStorage.GenerateDownloadURL(ctx, pin.Image_url, 1*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("generate download url for pin: %s: %w", pin.Id, err)
	}
	return &DownloadImgPinResponse{
		Pin:          *pin,
		Download_url: downloadURL,
	}, nil
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
