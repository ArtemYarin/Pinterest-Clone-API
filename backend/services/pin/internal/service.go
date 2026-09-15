package pin

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

var allowedImageContentTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
	"image/gif":  true,
}

const maxImageSizeBytes = 10 << 20 // 10MB

type PinService interface {
	CreatePin(ctx context.Context, userID uuid.UUID, pin CreatePinRequest) (*UploadImgPinResponse, error)
	GetPinByID(ctx context.Context, id string) (*DownloadImgPinResponse, error)
	GetPins(ctx context.Context, filters PinFilters) ([]*PinResponse, int, error)
	UpdatePin(ctx context.Context, id string, userID uuid.UUID, pin UpdatePinRequest) error
	DeletePin(ctx context.Context, id string, userID uuid.UUID) error
	ConfirmUpload(ctx context.Context, id string, userID uuid.UUID) (*PinResponse, error)
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

	// Download link - only for confirmed uploads, otherwise the object may not exist yet
	var downloadURL string
	if pin.Image_status == "confirmed" {
		downloadURL, err = s.imgStorage.GenerateDownloadURL(ctx, pin.Image_url, 1*time.Hour)
		if err != nil {
			return nil, fmt.Errorf("generate download url for pin: %s: %w", pin.Id, err)
		}
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

	// Best-effort cleanup - the DB row is already gone, so a storage failure here shouldn't fail the request
	if err := s.imgStorage.RemoveObject(ctx, existing.Image_url); err != nil {
		log.Printf("failed to remove image object %s for deleted pin %s: %v", existing.Image_url, id, err)
	}

	return nil
}

func (s *pinService) ConfirmUpload(ctx context.Context, id string, userID uuid.UUID) (*PinResponse, error) {
	// Ownership check
	existing, err := s.repo.GetPinByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get pin from repository: %w", err)
	}
	if existing.User_id != userID {
		return nil, fmt.Errorf("confirm upload: %w", errForbidden)
	}

	// Idempotent
	if existing.Image_status == "confirmed" {
		return existing, nil
	}

	info, err := s.imgStorage.StatObject(ctx, existing.Image_url)
	if err != nil {
		if isMinioNotFoundErr(err) {
			// Upload was never completed - drop the pin
			if delErr := s.repo.DeletePin(ctx, id); delErr != nil {
				log.Printf("failed to delete pin %s with missing upload: %v", id, delErr)
			}
			return nil, fmt.Errorf("stat image object: %w", errImageInvalid)
		}
		return nil, fmt.Errorf("stat image object: %w", errStorageUnavailable)
	}

	if !allowedImageContentTypes[info.ContentType] || info.Size > maxImageSizeBytes {
		if rmErr := s.imgStorage.RemoveObject(ctx, existing.Image_url); rmErr != nil {
			log.Printf("failed to remove invalid image object %s: %v", existing.Image_url, rmErr)
		}
		if delErr := s.repo.DeletePin(ctx, id); delErr != nil {
			log.Printf("failed to delete pin %s with invalid upload: %v", id, delErr)
		}
		return nil, fmt.Errorf("validate image (content-type: %s, size: %d): %w", info.ContentType, info.Size, errImageInvalid)
	}

	if err := s.repo.UpdateImageStatus(ctx, id, "confirmed"); err != nil {
		return nil, fmt.Errorf("update image status: %w", err)
	}

	existing.Image_status = "confirmed"
	return existing, nil
}
