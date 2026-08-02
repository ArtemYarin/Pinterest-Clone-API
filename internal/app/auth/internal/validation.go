package auth

import (
	"fmt"
	"net/mail"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/nbutton23/zxcvbn-go"
)

func PasswordStrengthValidation(password string) error {
	result := zxcvbn.PasswordStrength(password, nil)
	if result.Score < 3 {
		return newValidationErr(map[string]string{"password": "weak password"})
	}
	return nil
}
func EmailValidation(email string) error {
	_, err := mail.ParseAddress(email)
	if err != nil {
		valErr := newValidationErr(map[string]string{"email": "email must be a valid email address"})
		return fmt.Errorf("%w: %v", valErr, err)
	}
	return nil
}

func IsValidUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}

// Helpers
func formatValidationError(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", err.Field())
	case "email":
		return fmt.Sprintf("%s must be a valid email address", err.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s", err.Field(), err.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s", err.Field(), err.Param())
	default:
		return fmt.Sprintf("%s is invalid", err.Field())
	}
}

func getValidationMap(err error) map[string]string {
	errorsMap := make(map[string]string)
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, ve := range validationErrors {
			errorsMap[ve.Field()] = formatValidationError(ve)
		}
	}
	return errorsMap
}
