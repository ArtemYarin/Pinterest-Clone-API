package httpx

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteJSONError(t *testing.T) {
	errNotFound := NewError(http.StatusNotFound, "pin not found")

	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "wrapped typed error",
			err:        fmt.Errorf("get pin: %w", errNotFound),
			wantStatus: http.StatusNotFound,
			wantBody:   `{"code":"404","message":"pin not found"}`,
		},
		{
			name:       "common sentinel",
			err:        fmt.Errorf("decode: %w", ErrBadRequest),
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"code":"400","message":"Bad request"}`,
		},
		{
			name:       "validation error keeps Details key",
			err:        NewValidationErr(map[string]string{"email": "invalid"}),
			wantStatus: http.StatusUnprocessableEntity,
			wantBody:   `{"code":"422","Details":{"email":"invalid"}}`,
		},
		{
			name:       "unknown error",
			err:        errors.New("boom"),
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"code":"500","message":"Internal server error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			WriteJSONError(tt.err, rec)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
			assert.JSONEq(t, tt.wantBody, rec.Body.String())
		})
	}
}

func TestWriteJSON(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		rec := httptest.NewRecorder()
		WriteJSON(rec, http.StatusCreated, map[string]string{"id": "1"})

		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.JSONEq(t, `{"id":"1"}`, rec.Body.String())
	})

	t.Run("encode failure writes a single 500", func(t *testing.T) {
		rec := httptest.NewRecorder()
		WriteJSON(rec, http.StatusOK, make(chan int))

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.JSONEq(t, `{"code":"500","message":"Internal server error"}`, rec.Body.String())
	})
}
