package auth

import (
	"net/http"

	"github.com/ArtemYarin/pinterest-clone-api/pkg/httpx"
)

// Sentinel errors
var errEmailExists = httpx.NewError(http.StatusConflict, "Email arleady exists")
var errUserNotFound = httpx.NewError(http.StatusNotFound, "user not found")
var errTokenNotFound = httpx.NewError(http.StatusNotFound, "refresh token not found")
