package profile

import (
	"regexp"

	"github.com/ArtemYarin/pinterest-clone-api/pkg/httpx"
)

var usernameRegex = regexp.MustCompile(`^[a-z0-9_]{3,30}$`)

func UsernameValidation(username string) error {
	if !usernameRegex.MatchString(username) {
		return httpx.NewValidationErr(map[string]string{"username": "must be 3-30 lowercase letters, digits, or underscores"})
	}
	return nil
}
