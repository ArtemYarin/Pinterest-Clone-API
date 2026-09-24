package profile

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
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

const maxAvatarSizeBytes = 5 << 20 // 5MB

type ProfileService interface {
	CreateFromRegistration(ctx context.Context, userID uuid.UUID, suggestedUsername string) (*Profile, error)
	GetProfile(ctx context.Context, userID string) (*ProfileResponse, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, req UpdateProfileRequest) error
	GenerateAvatarUploadURL(ctx context.Context, userID uuid.UUID) (string, error)
	ConfirmAvatarUpload(ctx context.Context, userID uuid.UUID) (*ProfileResponse, error)
}

type profileService struct {
	repo       ProfileRepository
	validate   *validator.Validate
	imgStorage *ImageStorage
}

func NewProfileService(repo ProfileRepository, validate *validator.Validate, imgStorage *ImageStorage) ProfileService {
	return &profileService{repo: repo, validate: validate, imgStorage: imgStorage}
}

// CreateFromRegistration is called by the RabbitMQ consumer when auth-service
// publishes a user.registered event. It's idempotent: redelivery of the same
// event is safe (see profileRepository.CreateProfile).
func (s *profileService) CreateFromRegistration(ctx context.Context, userID uuid.UUID, suggestedUsername string) (*Profile, error) {
	username := sanitizeUsername(suggestedUsername)
	if username == "" {
		username = "user"
	}

	p, err := s.repo.CreateProfile(ctx, userID, username)
	if err != nil {
		return nil, fmt.Errorf("create profile in repo with userID: %v: %w", userID, err)
	}

	return p, err
}

func (s *profileService) GetProfile(ctx context.Context, userID string) (*ProfileResponse, error) {
	if !IsValidUUID(userID) {
		return nil, fmt.Errorf("get profile: %w", errBadRequest)
	}
	parsedID, _ := uuid.Parse(userID)

	p, err := s.repo.GetProfileByUserID(ctx, parsedID)
	if err != nil {
		return nil, fmt.Errorf("get profile from repository: %w", err)
	}

	var avatarURL string
	if p.AvatarStatus == "confirmed" && p.AvatarKey != nil {
		avatarURL, err = s.imgStorage.GenerateDownloadURL(ctx, *p.AvatarKey, 1*time.Hour)
		if err != nil {
			return nil, fmt.Errorf("generate avatar download url for profile %s: %w", p.UserID, err)
		}
	}

	return &ProfileResponse{
		UserID:    p.UserID,
		Username:  p.Username,
		Bio:       p.Bio,
		AvatarURL: avatarURL,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}, nil
}

func (s *profileService) UpdateProfile(ctx context.Context, userID uuid.UUID, req UpdateProfileRequest) error {
	err := s.validate.Struct(req)
	if err != nil {
		valErr := newValidationErr(getValidationMap(err))
		return fmt.Errorf("input validation: %w", valErr)
	}

	if req.Username != nil {
		if err := UsernameValidation(*req.Username); err != nil {
			return fmt.Errorf("username validation: %w", err)
		}
	}

	if req.Username == nil && req.Bio == nil {
		return fmt.Errorf("no fields to update: %w", errBadRequest)
	}

	if err := s.repo.UpdateProfile(ctx, userID, req.Username, req.Bio); err != nil {
		return fmt.Errorf("update profile in repository: %w", err)
	}
	return nil
}

func (s *profileService) GenerateAvatarUploadURL(ctx context.Context, userID uuid.UUID) (string, error) {
	// Ensure the profile exists before minting an upload URL for it
	if _, err := s.repo.GetProfileByUserID(ctx, userID); err != nil {
		return "", fmt.Errorf("get profile from repository: %w", err)
	}

	avatarKey := fmt.Sprintf("avatars/%s/%s", userID, uuid.New().String())
	uploadURL, err := s.imgStorage.GenerateUploadURL(ctx, avatarKey, 15*time.Minute)
	if err != nil {
		return "", fmt.Errorf("generate avatar upload URL: %w", err)
	}

	if err := s.repo.SetAvatar(ctx, userID, avatarKey, "pending"); err != nil {
		return "", fmt.Errorf("set avatar pending in repository: %w", err)
	}

	return uploadURL, nil
}

func (s *profileService) ConfirmAvatarUpload(ctx context.Context, userID uuid.UUID) (*ProfileResponse, error) {
	p, err := s.repo.GetProfileByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get profile from repository: %w", err)
	}
	if p.AvatarKey == nil {
		return nil, fmt.Errorf("confirm avatar: no pending upload: %w", errBadRequest)
	}

	// Idempotent
	if p.AvatarStatus == "confirmed" {
		return s.GetProfile(ctx, userID.String())
	}

	info, err := s.imgStorage.StatObject(ctx, *p.AvatarKey)
	if err != nil {
		if isMinioNotFoundErr(err) {
			return nil, fmt.Errorf("stat avatar object: %w", errImageInvalid)
		}
		return nil, fmt.Errorf("stat avatar object: %w", errStorageUnavailable)
	}

	if !allowedImageContentTypes[info.ContentType] || info.Size > maxAvatarSizeBytes {
		if rmErr := s.imgStorage.RemoveObject(ctx, *p.AvatarKey); rmErr != nil {
			log.Printf("failed to remove invalid avatar object %s: %v", *p.AvatarKey, rmErr)
		}
		return nil, fmt.Errorf("validate avatar (content-type: %s, size: %d): %w", info.ContentType, info.Size, errImageInvalid)
	}

	if err := s.repo.SetAvatar(ctx, userID, *p.AvatarKey, "confirmed"); err != nil {
		return nil, fmt.Errorf("update avatar status: %w", err)
	}

	return s.GetProfile(ctx, userID.String())
}

// sanitizeUsername lowercases raw and strips characters outside [a-z0-9_],
// truncating to 30 chars. Returns "" if the result is too short to be usable.
var usernameSanitizeRegex = regexp.MustCompile(`[^a-z0-9_]`)

func sanitizeUsername(raw string) string {
	cleaned := usernameSanitizeRegex.ReplaceAllString(strings.ToLower(raw), "")
	if len(cleaned) > 30 {
		cleaned = cleaned[:30]
	}
	if len(cleaned) < 3 {
		return ""
	}
	return cleaned
}
