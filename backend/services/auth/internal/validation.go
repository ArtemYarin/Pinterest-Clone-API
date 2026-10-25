package auth

import (
	"fmt"
	"net/mail"

	"github.com/ArtemYarin/pinterest-clone-api/pkg/httpx"
	"github.com/nbutton23/zxcvbn-go"
)

func PasswordStrengthValidation(password string) error {
	result := zxcvbn.PasswordStrength(password, nil)
	if result.Score < 3 {
		return httpx.NewValidationErr(map[string]string{"password": "weak password"})
	}
	return nil
}
func EmailValidation(email string) error {
	_, err := mail.ParseAddress(email)
	if err != nil {
		valErr := httpx.NewValidationErr(map[string]string{"email": "email must be a valid email address"})
		return fmt.Errorf("%w: %v", valErr, err)
	}
	return nil
}
