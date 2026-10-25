package pin

import (
	"net/http"

	"github.com/ArtemYarin/pinterest-clone-api/pkg/httpx"
	"github.com/minio/minio-go/v7"
)

// Sentinel errors
var errImageURLExists = httpx.NewError(http.StatusConflict, "Image URL already exists")
var errPinNotFound = httpx.NewError(http.StatusNotFound, "pin not found")
var errImageInvalid = httpx.NewError(http.StatusUnprocessableEntity, "Uploaded image is missing or failed validation")
var errStorageUnavailable = httpx.NewError(http.StatusBadGateway, "Image storage is temporarily unavailable")

// Helpers
func isMinioNotFoundErr(err error) bool {
	return minio.ToErrorResponse(err).Code == "NoSuchKey"
}
