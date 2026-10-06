package grpc

import (
	"context"
	"log"

	pinv1 "github.com/ArtemYarin/pinterest-clone-api/gen/pin/v1"
	pin "github.com/ArtemYarin/pinterest-clone-api/services/pin-service/internal"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	pinv1.UnimplementedPinServiceServer
	service pin.PinService
}

func NewServer(s pin.PinService) *Server {
	return &Server{service: s}
}

func (s *Server) GetPinsByIDs(ctx context.Context, req *pinv1.GetPinsByIDsRequest) (*pinv1.GetPinsByIDsResponse, error) {
	ids := make(uuid.UUIDs, 0, len(req.GetIds()))
	for _, rawID := range req.GetIds() {
		id, err := uuid.Parse(rawID)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid pin id %q", rawID)
		}
		ids = append(ids, id)
	}

	pins, err := s.service.GetPinsByIDs(ctx, ids)
	if err != nil {
		log.Printf("grpc get pins by ids (%d ids): %v", len(ids), err)
		return nil, status.Error(codes.Internal, "get pins by ids")
	}

	resp := &pinv1.GetPinsByIDsResponse{Pins: make([]*pinv1.Pin, 0, len(pins))}
	for _, p := range pins {
		resp.Pins = append(resp.Pins, toProto(p))
	}
	return resp, nil
}

func toProto(p pin.DownloadImgPinResponse) *pinv1.Pin {
	return &pinv1.Pin{
		Id:          p.Pin.Id.String(),
		UserId:      p.Pin.User_id.String(),
		Title:       p.Pin.Title,
		ImageUrl:    p.Pin.Image_url,
		ImageStatus: p.Pin.Image_status,
		Description: p.Pin.Description,
		CreatedAt:   timestamppb.New(p.Pin.Created_at),
		UpdatedAt:   timestamppb.New(p.Pin.Updated_at),
		Likes:       int64(p.Pin.Likes),
		DownloadUrl: p.Download_url,
	}
}
