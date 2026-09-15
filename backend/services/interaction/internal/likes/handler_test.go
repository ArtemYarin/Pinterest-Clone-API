package likes

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ArtemYarin/pinterest-clone-api/pkg/jwt"
	"github.com/ArtemYarin/pinterest-clone-api/pkg/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// newLikeRequest builds a request with the chi "pinID" URL param (when pinIDStr
// is non-empty) and, when claims is non-nil, authenticated user claims attached.
func newLikeRequest(method, target, pinIDStr string, claims *jwt.Claims) *http.Request {
	req := httptest.NewRequest(method, target, nil)

	rctx := chi.NewRouteContext()
	if pinIDStr != "" {
		rctx.URLParams.Add("pinID", pinIDStr)
	}
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)

	if claims != nil {
		ctx = context.WithValue(ctx, middleware.UserKey, claims)
	}

	return req.WithContext(ctx)
}

func TestLikeHandler_GetLikeCount(t *testing.T) {
	pinID := uuid.New()

	t.Run("invalid pinID", func(t *testing.T) {
		svc := new(mockLikeService)
		h := NewLikeHandler(svc)
		req := newLikeRequest(http.MethodGet, "/x/like/count", "not-a-uuid", nil)
		w := httptest.NewRecorder()

		h.GetLikeCount(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
		svc.AssertNotCalled(t, "GetLikeCount", mock.Anything, mock.Anything)
	})

	t.Run("service error", func(t *testing.T) {
		svc := new(mockLikeService)
		svc.On("GetLikeCount", mock.Anything, pinID).Return((*int64)(nil), errors.New("boom"))
		h := NewLikeHandler(svc)
		req := newLikeRequest(http.MethodGet, "/x/like/count", pinID.String(), nil)
		w := httptest.NewRecorder()

		h.GetLikeCount(w, req)

		require.Equal(t, http.StatusInternalServerError, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		svc := new(mockLikeService)
		count := int64(7)
		svc.On("GetLikeCount", mock.Anything, pinID).Return(&count, nil)
		h := NewLikeHandler(svc)
		req := newLikeRequest(http.MethodGet, "/x/like/count", pinID.String(), nil)
		w := httptest.NewRecorder()

		h.GetLikeCount(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		require.JSONEq(t, `{"pin_id":"`+pinID.String()+`","count":7}`, w.Body.String())
		svc.AssertExpectations(t)
	})
}

func TestLikeHandler_HasLiked(t *testing.T) {
	pinID := uuid.New()
	userID := uuid.New()
	claims := &jwt.Claims{UserID: userID}

	t.Run("invalid pinID", func(t *testing.T) {
		svc := new(mockLikeService)
		h := NewLikeHandler(svc)
		req := newLikeRequest(http.MethodGet, "/x/like", "not-a-uuid", claims)
		w := httptest.NewRecorder()

		h.HasLiked(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing claims", func(t *testing.T) {
		svc := new(mockLikeService)
		h := NewLikeHandler(svc)
		req := newLikeRequest(http.MethodGet, "/x/like", pinID.String(), nil)
		w := httptest.NewRecorder()

		h.HasLiked(w, req)

		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		svc := new(mockLikeService)
		svc.On("HasLiked", mock.Anything, userID, pinID).Return((*bool)(nil), errors.New("boom"))
		h := NewLikeHandler(svc)
		req := newLikeRequest(http.MethodGet, "/x/like", pinID.String(), claims)
		w := httptest.NewRecorder()

		h.HasLiked(w, req)

		require.Equal(t, http.StatusInternalServerError, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		svc := new(mockLikeService)
		liked := true
		svc.On("HasLiked", mock.Anything, userID, pinID).Return(&liked, nil)
		h := NewLikeHandler(svc)
		req := newLikeRequest(http.MethodGet, "/x/like", pinID.String(), claims)
		w := httptest.NewRecorder()

		h.HasLiked(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		require.JSONEq(t, `{"pin_id":"`+pinID.String()+`","liked":true}`, w.Body.String())
		svc.AssertExpectations(t)
	})
}

func TestLikeHandler_HasLikedBatch(t *testing.T) {
	userID := uuid.New()
	claims := &jwt.Claims{UserID: userID}
	pinIDs := uuid.UUIDs{uuid.New(), uuid.New()}

	t.Run("missing pin_ids", func(t *testing.T) {
		svc := new(mockLikeService)
		h := NewLikeHandler(svc)
		req := newLikeRequest(http.MethodGet, "/likes/batch", "", claims)
		w := httptest.NewRecorder()

		h.HasLikedBatch(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid pin_ids", func(t *testing.T) {
		svc := new(mockLikeService)
		h := NewLikeHandler(svc)
		req := newLikeRequest(http.MethodGet, "/likes/batch?pin_ids=not-a-uuid", "", claims)
		w := httptest.NewRecorder()

		h.HasLikedBatch(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing claims", func(t *testing.T) {
		svc := new(mockLikeService)
		h := NewLikeHandler(svc)
		req := newLikeRequest(http.MethodGet, "/likes/batch?pin_ids="+pinIDs[0].String(), "", nil)
		w := httptest.NewRecorder()

		h.HasLikedBatch(w, req)

		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		svc := new(mockLikeService)
		svc.On("HasLikedBatch", mock.Anything, userID, pinIDs).Return(map[uuid.UUID]bool(nil), errors.New("boom"))
		h := NewLikeHandler(svc)
		target := "/likes/batch?pin_ids=" + pinIDs[0].String() + "," + pinIDs[1].String()
		req := newLikeRequest(http.MethodGet, target, "", claims)
		w := httptest.NewRecorder()

		h.HasLikedBatch(w, req)

		require.Equal(t, http.StatusInternalServerError, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		svc := new(mockLikeService)
		expected := map[uuid.UUID]bool{pinIDs[0]: true, pinIDs[1]: false}
		svc.On("HasLikedBatch", mock.Anything, userID, pinIDs).Return(expected, nil)
		h := NewLikeHandler(svc)
		target := "/likes/batch?pin_ids=" + pinIDs[0].String() + "+,+" + pinIDs[1].String()
		req := newLikeRequest(http.MethodGet, target, "", claims)
		w := httptest.NewRecorder()

		h.HasLikedBatch(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})
}

func TestLikeHandler_ListLikedByUser(t *testing.T) {
	userID := uuid.New()
	claims := &jwt.Claims{UserID: userID}

	t.Run("missing claims", func(t *testing.T) {
		svc := new(mockLikeService)
		h := NewLikeHandler(svc)
		req := newLikeRequest(http.MethodGet, "/likes", "", nil)
		w := httptest.NewRecorder()

		h.ListLikedByUser(w, req)

		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		svc := new(mockLikeService)
		svc.On("ListLikedByUser", mock.Anything, userID).Return([]uuid.UUID(nil), errors.New("boom"))
		h := NewLikeHandler(svc)
		req := newLikeRequest(http.MethodGet, "/likes", "", claims)
		w := httptest.NewRecorder()

		h.ListLikedByUser(w, req)

		require.Equal(t, http.StatusInternalServerError, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		svc := new(mockLikeService)
		pinIDs := []uuid.UUID{uuid.New(), uuid.New()}
		svc.On("ListLikedByUser", mock.Anything, userID).Return(pinIDs, nil)
		h := NewLikeHandler(svc)
		req := newLikeRequest(http.MethodGet, "/likes", "", claims)
		w := httptest.NewRecorder()

		h.ListLikedByUser(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})
}
