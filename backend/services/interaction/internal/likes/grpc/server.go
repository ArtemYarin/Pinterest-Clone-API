package grpc

import (
	"context"
	"log"

	likesv1 "github.com/ArtemYarin/pinterest-clone-api/gen/likes/v1"
	"github.com/ArtemYarin/pinterest-clone-api/services/interaction-service/internal/likes"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	likesv1.UnimplementedLikesServiceServer
	service likes.LikeService
}

func NewServer(s likes.LikeService) *Server {
	return &Server{service: s}
}

func (s *Server) ListLikedPinIDs(ctx context.Context, req *likesv1.ListLikedPinIDsRequest) (*likesv1.ListLikedPinIDsResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user id %q", req.GetUserId())
	}

	pinIDs, err := s.service.ListLikedByUser(ctx, userID)
	if err != nil {
		log.Printf("grpc list liked pin ids user %v: %v", userID, err)
		return nil, status.Error(codes.Internal, "list liked pin ids")
	}

	resp := &likesv1.ListLikedPinIDsResponse{PinIds: make([]string, 0, len(pinIDs))}
	for _, id := range pinIDs {
		resp.PinIds = append(resp.PinIds, id.String())
	}
	return resp, nil
}
