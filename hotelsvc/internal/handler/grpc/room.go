package grpchandler

import (
	"context"
	roomsv1 "github.com/olegetoya/booking/protos/gen/go/hotelsvc/rooms"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Rooms interface {
	GetAll(ctx context.Context, hotelID int64) ([]*roomsv1.Room, error)
	Get(ctx context.Context, hotelId, roomID int64) (*roomsv1.Room, error)
}

type serverAPI struct {
	roomsv1.UnimplementedRoomsServer
	rooms Rooms
}

func Register(gRPC *grpc.Server, rooms Rooms) {
	roomsv1.RegisterRoomsServer(gRPC, &serverAPI{rooms: rooms})
}

func (r *serverAPI) GetAll(ctx context.Context, request *roomsv1.GetAllRequest) (*roomsv1.GetAllResponse, error) {
	err := validateGetAll(request)
	if err != nil {
		return nil, err
	}
	rooms, err := r.rooms.GetAll(ctx, request.HotelID)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}
	return &roomsv1.GetAllResponse{Rooms: rooms}, nil
}

func (r *serverAPI) Get(ctx context.Context, request *roomsv1.GetRequest) (*roomsv1.GetResponse, error) {
	err := validateGet(request)
	if err != nil {
		return nil, err
	}
	room, err := r.rooms.Get(ctx, request.HotelID, request.RoomID)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}
	return &roomsv1.GetResponse{Room: room}, nil
}

func validateGetAll(r *roomsv1.GetAllRequest) error {
	if r.HotelID < 1 {
		return status.Error(codes.InvalidArgument, "hotel ID should be greater than zero")
	}
	return nil
}

func validateGet(r *roomsv1.GetRequest) error {
	if r.HotelID < 1 {
		return status.Error(codes.InvalidArgument, "hotel ID should be greater than zero")
	}
	if r.RoomID < 1 {
		return status.Error(codes.InvalidArgument, "room ID should be greater than zero")
	}
	return nil
}
