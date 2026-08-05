package pin

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

func IsValidUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}

// Helpers
func formatValidationError(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", err.Field())
	case "url":
		return fmt.Sprintf("%s must be a valid URL", err.Field())
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
