package users

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	likesv1 "github.com/ArtemYarin/pinterest-clone-api/gen/likes/v1"
	pinv1 "github.com/ArtemYarin/pinterest-clone-api/gen/pin/v1"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const likedPinsTimeout = 5 * time.Second

type LikedPinsHandler struct {
	likes likesv1.LikesServiceClient
	pins  pinv1.PinServiceClient
}

func NewLikedPinsHandler(likes likesv1.LikesServiceClient, pins pinv1.PinServiceClient) *LikedPinsHandler {
	return &LikedPinsHandler{likes: likes, pins: pins}
}

// Response shape mirrors pin-service GET /pin so the frontend can reuse it
type pinResponse struct {
	Id           string    `json:"id"`
	User_id      string    `json:"user_id"`
	Title        string    `json:"title"`
	Image_url    string    `json:"image_url"`
	Image_status string    `json:"image_status"`
	Description  *string   `json:"description,omitempty"`
	Created_at   time.Time `json:"created_at"`
	Updated_at   time.Time `json:"updated_at"`
	Likes        int64     `json:"likes"`
}

type downloadImgPinResponse struct {
	Pin          pinResponse `json:"pin"`
	Download_url string      `json:"download_url"`
}

// GetLikedPins composes likes and pin services: it fetches IDs of pins liked by
// the user, then their data, and returns them in like order (newest first)
func (h *LikedPinsHandler) GetLikedPins(w http.ResponseWriter, r *http.Request) {
	// UUID parse
	userID, err := uuid.Parse(chi.URLParam(r, "userID"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), likedPinsTimeout)
	defer cancel()

	// Get pin IDs from interactions service
	liked, err := h.likes.ListLikedPinIDs(ctx, &likesv1.ListLikedPinIDsRequest{UserId: userID.String()})
	if err != nil {
		log.Printf("get liked pins: list liked pin ids user %v: %v", userID, err)
		writeGRPCError(w, err)
		return
	}

	data := []downloadImgPinResponse{}

	if len(liked.GetPinIds()) > 0 {
		// Get actual pins by IDs (object)
		resp, err := h.pins.GetPinsByIDs(ctx, &pinv1.GetPinsByIDsRequest{Ids: liked.GetPinIds()})
		if err != nil {
			log.Printf("get liked pins: get pins by ids user %v: %v", userID, err)
			writeGRPCError(w, err)
			return
		}

		// Unwrap and reorganize gRPC message.
		byID := make(map[string]*pinv1.Pin, len(resp.GetPins()))
		for _, p := range resp.GetPins() {
			byID[p.GetId()] = p
		}

		// Keep like order, skip pins that were deleted or aren't confirmed
		for _, id := range liked.GetPinIds() {
			if p, ok := byID[id]; ok {
				data = append(data, fromProto(p))
			}
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": data,
		"metadata": map[string]any{
			"total":   len(data),
			"user_id": userID,
		},
	})
}

// Receives: proto Pin.
// Returns: downloadImgPinResponse models
func fromProto(p *pinv1.Pin) downloadImgPinResponse {
	return downloadImgPinResponse{
		Pin: pinResponse{
			Id:           p.GetId(),
			User_id:      p.GetUserId(),
			Title:        p.GetTitle(),
			Image_url:    p.GetImageUrl(),
			Image_status: p.GetImageStatus(),
			Description:  p.Description,
			Created_at:   p.GetCreatedAt().AsTime(),
			Updated_at:   p.GetUpdatedAt().AsTime(),
			Likes:        p.GetLikes(),
		},
		Download_url: p.GetDownloadUrl(),
	}
}

// Helpers
func writeGRPCError(w http.ResponseWriter, err error) {
	switch status.Code(err) {
	case codes.InvalidArgument:
		writeJSONError(w, http.StatusBadRequest, "Bad request")
	case codes.Unavailable:
		writeJSONError(w, http.StatusServiceUnavailable, "Service unavailable")
	case codes.DeadlineExceeded:
		writeJSONError(w, http.StatusGatewayTimeout, "Upstream timeout")
	default:
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
	}
}

func writeJSONError(w http.ResponseWriter, code int, message string) {
	writeJSON(w, code, map[string]string{
		"code":    strconv.Itoa(code),
		"message": message,
	})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("write json: %v", err)
	}
}
