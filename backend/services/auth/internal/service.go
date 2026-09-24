package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/ArtemYarin/pinterest-clone-api/pkg/jwt"
	"github.com/ArtemYarin/pinterest-clone-api/pkg/rabbitmq"
	"github.com/ArtemYarin/pinterest-clone-api/services/auth-service/internal/password"
	"github.com/go-playground/validator/v10"
	rmq "github.com/rabbitmq/rabbitmq-amqp-go-client/pkg/rabbitmqamqp"
)

// EventPublisher is the subset of *rabbitmq.Publisher the auth service needs,
// kept as an interface so it can be swapped/mocked like UserRepository.
type EventPublisher interface {
	Publish(ctx context.Context, routingKey string, payload any) error
}

type UserService interface {
	RegisterUser(ctx context.Context, user CredentialsUserRequest) (*UserWithTokenResponse, error)
	LoginUser(ctx context.Context, user CredentialsUserRequest) (string, error)
	GetUserByEmail(ctx context.Context, email string) (*UserResponse, error)
	GetUserByID(ctx context.Context, id string) (*UserResponse, error)
	UpdateUser(ctx context.Context, user UpdateUserRequest) error
}

type userService struct {
	repo      UserRepository
	validate  *validator.Validate
	publisher *rmq.Publisher
}

func NewUserService(repo UserRepository, validate *validator.Validate, publisher *rmq.Publisher) UserService {
	return &userService{repo: repo, validate: validate, publisher: publisher}
}

func (s *userService) RegisterUser(ctx context.Context, user CredentialsUserRequest) (*UserWithTokenResponse, error) {
	// Validate input
	err := s.validate.Struct(user)
	if err != nil {
		valErr := newValidationErr(getValidationMap(err))
		return nil, fmt.Errorf("input validation: %w", valErr)
	}

	// Validate password
	err = PasswordStrengthValidation(user.Password_hash)
	if err != nil {
		return nil, fmt.Errorf("password validation: %w", err)
	}

	// Password hashing
	hashedPassword, err := password.HashPassword(user.Password_hash)
	if err != nil {
		return nil, fmt.Errorf("password hashing %s: %w", user.Password_hash, err)
	}
	user.Password_hash = hashedPassword

	// Repository call
	u, err := s.repo.CreateUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("create user in repository: %w", err)
	}

	// Token
	token, err := jwt.GenerateToken(u.Id)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w:", err)
	}

	// Publish registration event for profile-service to consume.
	// A publish failure shouldn't fail signup, since the two services are
	// meant to stay loosely coupled.
	if s.publisher != nil {
		evt := rabbitmq.UserRegisteredEvent{ // pkg dependecy
			UserID:            u.Id,
			Email:             u.Email,
			SuggestedUsername: deriveUsernameFromEmail(u.Email),
			OccurredAt:        time.Now().UTC(),
		}
		body, err := json.Marshal(evt)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal user.registered event for user %s: %v", u.Id, err)
		}
		res, err := s.publisher.Publish(ctx, rmq.NewMessage(body))
		if err != nil {
			return nil, fmt.Errorf("failed to publish user.registered event for user %s: %v", u.Id, err)
		}
		switch res.Outcome.(type) {
		case *rmq.StateAccepted:
		default:
			return nil, fmt.Errorf("unexpected publish outcome: %v", res.Outcome)
		}
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
		valErr := newValidationErr(getValidationMap(err))
		return "", fmt.Errorf("service LoginUser: %w", valErr)
	}

	// Repo
	storedUser, err := s.repo.GetUserByEmail(ctx, user.Email)
	if err != nil {
		return "", fmt.Errorf("get user from repository: %w", err)
	}

	// Password validation
	if err := password.CheckPassword(storedUser.Password_hash, user.Password_hash); err != nil {
		return "", fmt.Errorf("compare password: %w", errUnauthorized)
	}

	// Token
	token, err := jwt.GenerateToken(storedUser.Id)
	if err != nil {
		return "", fmt.Errorf("generate token: %w:", err)
	}

	return token, nil
}

func (s *userService) GetUserByEmail(ctx context.Context, email string) (*UserResponse, error) {
	err := EmailValidation(email)
	if err != nil {
		return nil, fmt.Errorf("email validation: %w", err)
	}
	userData, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("get user from repository: %w", err)
	}
	return &UserResponse{
		Id:         userData.Id,
		Email:      userData.Email,
		Created_at: userData.Created_at,
		Updated_at: userData.Updated_at,
	}, nil
}

func (s *userService) GetUserByID(ctx context.Context, id string) (*UserResponse, error) {
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user from repository: %w", err)
	}
	return user, nil
}

func (s *userService) UpdateUser(ctx context.Context, user UpdateUserRequest) error {
	// Validate input
	err := s.validate.Struct(user)
	if err != nil {
		valErr := newValidationErr(getValidationMap(err))
		return fmt.Errorf("service UpdateUser: %w", valErr)
	}

	// Validate and hash password if provided
	if user.Password_hash != nil {
		err = PasswordStrengthValidation(*user.Password_hash)
		if err != nil {
			return fmt.Errorf("password strength validation: %w", err)
		}

		hashedPassword, err := password.HashPassword(*user.Password_hash)
		if err != nil {
			return fmt.Errorf("hashing password: %w", err)
		}
		*user.Password_hash = hashedPassword
	}

	err = s.repo.UpdateUser(ctx, user)
	if err != nil {
		return fmt.Errorf("update user in repository: %w", err)
	}
	return nil
}

var usernameSanitizeRegex = regexp.MustCompile(`[^a-z0-9_]`)

// deriveUsernameFromEmail builds a default profile username from the local
// part of an email address (e.g. "artem" from "artem@example.com").
func deriveUsernameFromEmail(email string) string {
	local := email
	if i := strings.Index(email, "@"); i >= 0 {
		local = email[:i]
	}

	cleaned := usernameSanitizeRegex.ReplaceAllString(strings.ToLower(local), "")
	if len(cleaned) > 30 {
		cleaned = cleaned[:30]
	}
	if len(cleaned) < 3 {
		cleaned = "user"
	}
	return cleaned
}
