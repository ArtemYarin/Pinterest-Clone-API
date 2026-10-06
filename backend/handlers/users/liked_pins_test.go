package users

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	likesv1 "github.com/ArtemYarin/pinterest-clone-api/gen/likes/v1"
	pinv1 "github.com/ArtemYarin/pinterest-clone-api/gen/pin/v1"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockLikesClient struct {
	mock.Mock
}

func (m *mockLikesClient) ListLikedPinIDs(ctx context.Context, in *likesv1.ListLikedPinIDsRequest, opts ...grpc.CallOption) (*likesv1.ListLikedPinIDsResponse, error) {
	args := m.Called(ctx, in)
	resp, _ := args.Get(0).(*likesv1.ListLikedPinIDsResponse)
	return resp, args.Error(1)
}

type mockPinsClient struct {
	mock.Mock
}

func (m *mockPinsClient) GetPinsByIDs(ctx context.Context, in *pinv1.GetPinsByIDsRequest, opts ...grpc.CallOption) (*pinv1.GetPinsByIDsResponse, error) {
	args := m.Called(ctx, in)
	resp, _ := args.Get(0).(*pinv1.GetPinsByIDsResponse)
	return resp, args.Error(1)
}

func serve(h *LikedPinsHandler, userID string) *httptest.ResponseRecorder {
	r := chi.NewRouter()
	r.Get("/users/{userID}/liked-pins", h.GetLikedPins)

	req := httptest.NewRequest(http.MethodGet, "/users/"+userID+"/liked-pins", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

type likedPinsBody struct {
	Data     []downloadImgPinResponse `json:"data"`
	Metadata struct {
		Total int `json:"total"`
	} `json:"metadata"`
}

func TestLikedPinsHandler_GetLikedPins(t *testing.T) {
	userID := uuid.New()
	first, second, deleted := uuid.NewString(), uuid.NewString(), uuid.NewString()

	t.Run("returns pins in like order and skips missing ones", func(t *testing.T) {
		likes := new(mockLikesClient)
		pins := new(mockPinsClient)
		ids := []string{first, deleted, second}

		likes.On("ListLikedPinIDs", mock.Anything, &likesv1.ListLikedPinIDsRequest{UserId: userID.String()}).
			Return(&likesv1.ListLikedPinIDsResponse{PinIds: ids}, nil)
		// Pin service returns them in a different order and without the deleted one
		pins.On("GetPinsByIDs", mock.Anything, &pinv1.GetPinsByIDsRequest{Ids: ids}).
			Return(&pinv1.GetPinsByIDsResponse{Pins: []*pinv1.Pin{
				{Id: second, Title: "second", DownloadUrl: "url-2"},
				{Id: first, Title: "first", DownloadUrl: "url-1"},
			}}, nil)

		rec := serve(NewLikedPinsHandler(likes, pins), userID.String())

		require.Equal(t, http.StatusOK, rec.Code)
		var body likedPinsBody
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		require.Len(t, body.Data, 2)
		assert.Equal(t, first, body.Data[0].Pin.Id)
		assert.Equal(t, "url-1", body.Data[0].Download_url)
		assert.Equal(t, second, body.Data[1].Pin.Id)
		assert.Equal(t, 2, body.Metadata.Total)
		likes.AssertExpectations(t)
		pins.AssertExpectations(t)
	})

	t.Run("no likes skips pin service call", func(t *testing.T) {
		likes := new(mockLikesClient)
		pins := new(mockPinsClient)

		likes.On("ListLikedPinIDs", mock.Anything, mock.Anything).
			Return(&likesv1.ListLikedPinIDsResponse{}, nil)

		rec := serve(NewLikedPinsHandler(likes, pins), userID.String())

		require.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `[]`, mustField(t, rec, "data"))
		pins.AssertNotCalled(t, "GetPinsByIDs", mock.Anything, mock.Anything)
	})

	t.Run("invalid user id returns 400", func(t *testing.T) {
		likes := new(mockLikesClient)
		pins := new(mockPinsClient)

		rec := serve(NewLikedPinsHandler(likes, pins), "not-a-uuid")

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		likes.AssertNotCalled(t, "ListLikedPinIDs", mock.Anything, mock.Anything)
	})

	t.Run("likes service unavailable returns 503", func(t *testing.T) {
		likes := new(mockLikesClient)
		pins := new(mockPinsClient)

		likes.On("ListLikedPinIDs", mock.Anything, mock.Anything).
			Return(nil, status.Error(codes.Unavailable, "down"))

		rec := serve(NewLikedPinsHandler(likes, pins), userID.String())

		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
		pins.AssertNotCalled(t, "GetPinsByIDs", mock.Anything, mock.Anything)
	})

	t.Run("pin service internal error returns 500", func(t *testing.T) {
		likes := new(mockLikesClient)
		pins := new(mockPinsClient)

		likes.On("ListLikedPinIDs", mock.Anything, mock.Anything).
			Return(&likesv1.ListLikedPinIDsResponse{PinIds: []string{first}}, nil)
		pins.On("GetPinsByIDs", mock.Anything, mock.Anything).
			Return(nil, status.Error(codes.Internal, "boom"))

		rec := serve(NewLikedPinsHandler(likes, pins), userID.String())

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func mustField(t *testing.T, rec *httptest.ResponseRecorder, field string) string {
	t.Helper()
	var body map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	return string(body[field])
}
