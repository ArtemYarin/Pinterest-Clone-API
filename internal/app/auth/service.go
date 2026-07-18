package auth

import (
	"context"
	"fmt"

	"github.com/ArtemYarin/pinterest-clone-api/internal/app/auth/password"
	"github.com/ArtemYarin/pinterest-clone-api/internal/app/auth/validation"
	"github.com/ArtemYarin/pinterest-clone-api/pkg/jwt"
	"github.com/go-playground/validator/v10"
)

type UserService interface {
	RegisterUser(ctx context.Context, user CredentialsUserRequest) (*UserWithTokenResponse, error)
	LoginUser(ctx context.Context, user CredentialsUserRequest) (string, error)
	GetUserByEmail(ctx context.Context, email string) (*UserWithPasswordResponse, error)
	GetUserByID(ctx context.Context, id string) (*UserWithPasswordResponse, error)
	UpdateUser(ctx context.Context, user UpdateUserRequest) error
}

type userService struct {
	repo     UserRepository
	validate *validator.Validate
}

func NewUserService(repo UserRepository, validate *validator.Validate) UserService {
	return &userService{repo: repo, validate: validate}
}

func (s *userService) RegisterUser(ctx context.Context, user CredentialsUserRequest) (*UserWithTokenResponse, error) {
	// Validate input
	err := s.validate.Struct(user)
	if err != nil {
		return nil, fmt.Errorf("service RegisterUser: %v", err)
	}

	// Validate password
	err = validation.PasswordStrengthValidation(user.Password_hash)
	if err != nil {
		return nil, fmt.Errorf("service RegisterUser: %v", err)
	}

	// Password hashing
	hashedPassword, err := password.HashPassword(user.Password_hash)
	if err != nil {
		return nil, fmt.Errorf("service RegisterUser: %v", err)
	}
	user.Password_hash = hashedPassword

	// Repo
	u, err := s.repo.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	// Token
	token, err := jwt.GenerateToken(u.Id)
	if err != nil {
		return nil, err
	}

	return &UserWithTokenResponse{
		UserResponse: *u,
		Token:        token,
	}, nil
}

func (s *userService) LoginUser(ctx context.Context, user CredentialsUserRequest) (string, error) {
	// Validate input
	err := s.validate.Struct(user)
	if err != nil {
		return "", fmt.Errorf("service LoginUser: %v", err)
	}

	// Repo
	storedUser, err := s.repo.GetUserByEmail(ctx, user.Email)
	if err != nil {
		return "", err
	}

	// Password validation
	if err := password.CheckPassword(user.Password_hash, storedUser.Password_hash); err != nil {
		return "", fmt.Errorf("service LoginUser: invalid credentials")
	}

	// Token
	token, err := jwt.GenerateToken(storedUser.Id)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *userService) GetUserByEmail(ctx context.Context, email string) (*UserWithPasswordResponse, error) {
	err := validation.EmailValidation(email)
	if err != nil {
		return nil, err
	}
	return s.repo.GetUserByEmail(ctx, email)
}

func (s *userService) GetUserByID(ctx context.Context, id string) (*UserWithPasswordResponse, error) {
	return s.repo.GetUserByID(ctx, id)
}

func (s *userService) UpdateUser(ctx context.Context, user UpdateUserRequest) error {
	// Validate input
	err := s.validate.Struct(user)
	if err != nil {
		return fmt.Errorf("service RegisterUser: %v", err)
	}

	return s.repo.UpdateUser(ctx, user)
}
