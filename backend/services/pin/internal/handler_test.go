package pin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ArtemYarin/pinterest-clone-api/pkg/jwt"
	"github.com/ArtemYarin/pinterest-clone-api/pkg/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// newPinRequest builds a request with the chi "id" URL param (when id is
// non-empty) and, when claims is non-nil, authenticated user claims attached.
func newPinRequest(method, target, body, id string, claims *jwt.Claims) *http.Request {
	req := httptest.NewRequest(method, target, strings.NewReader(body))

	rctx := chi.NewRouteContext()
	if id != "" {
		rctx.URLParams.Add("id", id)
	}
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)

	if claims != nil {
		ctx = context.WithValue(ctx, middleware.UserKey, claims)
	}

	return req.WithContext(ctx)
}

func newTestHandler(t *testing.T) (*PinHandler, *mockPinService) {
	t.Helper()
	svc := new(mockPinService)
	svc.Test(t)
	t.Cleanup(func() { svc.AssertExpectations(t) })
	h := NewPinHandler(svc)
	return &h, svc
}

func TestPinHandler_CreatePin(t *testing.T) {
	claims := &jwt.Claims{UserID: uuid.New()}

	tests := []struct {
		name       string
		body       string
		claims     *jwt.Claims
		setup      func(svc *mockPinService)
		wantStatus int
	}{
		{
			name:       "malformed json",
			body:       `{"title":`,
			claims:     claims,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "unknown field",
			body:       `{"title":"Sunset","owner":"me"}`,
			claims:     claims,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "unauthenticated",
			body:       `{"title":"Sunset"}`,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:   "validation error",
			body:   `{"title":""}`,
			claims: claims,
			setup: func(svc *mockPinService) {
				svc.On("CreatePin", mock.Anything, claims.UserID, CreatePinRequest{Title: ""}).
					Return(nil, newValidationErr(map[string]string{"Title": "Title is required"}))
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:   "success",
			body:   `{"title":"Sunset"}`,
			claims: claims,
			setup: func(svc *mockPinService) {
				svc.On("CreatePin", mock.Anything, claims.UserID, CreatePinRequest{Title: "Sunset"}).
					Return(&UploadImgPinResponse{Upload_url: "https://upload.example"}, nil)
			},
			wantStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, svc := newTestHandler(t)
			if tt.setup != nil {
				tt.setup(svc)
			}
			w := httptest.NewRecorder()

			h.CreatePin(w, newPinRequest(http.MethodPost, "/", tt.body, "", tt.claims))

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		})
	}
}

func TestPinHandler_GetPinByID(t *testing.T) {
	pinID := uuid.New()

	tests := []struct {
		name       string
		id         string
		setup      func(svc *mockPinService)
		wantStatus int
	}{
		{
			name:       "invalid uuid",
			id:         "not-a-uuid",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "not found",
			id:   pinID.String(),
			setup: func(svc *mockPinService) {
				svc.On("GetPinByID", mock.Anything, pinID.String()).Return(nil, errPinNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "success",
			id:   pinID.String(),
			setup: func(svc *mockPinService) {
				svc.On("GetPinByID", mock.Anything, pinID.String()).
					Return(&DownloadImgPinResponse{Pin: PinResponse{Id: pinID}}, nil)
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, svc := newTestHandler(t)
			if tt.setup != nil {
				tt.setup(svc)
			}
			w := httptest.NewRecorder()

			h.GetPinByID(w, newPinRequest(http.MethodGet, "/"+tt.id, "", tt.id, nil))

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestPinHandler_GetPins_ServiceError(t *testing.T) {
	h, svc := newTestHandler(t)
	svc.On("GetPins", mock.Anything, mock.Anything).Return(nil, 0, errInternalServer)
	w := httptest.NewRecorder()

	h.GetPins(w, newPinRequest(http.MethodGet, "/", "", "", nil))

	require.Equal(t, http.StatusInternalServerError, w.Code)

	// The body must be exactly one JSON error object - no trailing success payload
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body), "body: %s", w.Body.String())
	assert.Equal(t, "500", body["code"])
}

func TestWriteJSONError(t *testing.T) {
	tests := []struct {
		err        error
		wantStatus int
	}{
		{newValidationErr(map[string]string{"Title": "Title is required"}), http.StatusUnprocessableEntity},
		{errBadRequest, http.StatusBadRequest},
		{errUnauthorized, http.StatusUnauthorized},
		{errForbidden, http.StatusForbidden},
		{errPinNotFound, http.StatusNotFound},
		{errImageURLExists, http.StatusConflict},
		{errImageInvalid, http.StatusUnprocessableEntity},
		{errStorageUnavailable, http.StatusBadGateway},
		{errors.New("unexpected"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.err.Error(), func(t *testing.T) {
			w := httptest.NewRecorder()

			// Wrapped like the handlers do, so the mapping must unwrap the chain
			WriteJSONError(fmt.Errorf("handler: %w", tt.err), w)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
