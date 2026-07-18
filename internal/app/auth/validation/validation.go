package validation

import (
	"errors"
	"net/mail"

	"github.com/nbutton23/zxcvbn-go"
)

func PasswordStrengthValidation(password string) error {
	result := zxcvbn.PasswordStrength(password, nil)
	if result.Score < 3 {
		return errors.New("weak password")
	}
	return nil
}
func EmailValidation(email string) error {
	_, err := mail.ParseAddress(email)
	return err
}
