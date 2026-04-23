package grpc

import (
	"context"
	"github.com/olegetoya/booking/hotelsvc/internal/dto"
	roomsv1 "github.com/olegetoya/booking/protos/gen/go/hotelsvc/rooms"
	"google.golang.org/grpc"
)

type Rooms interface {
	GetAll(hotelID int64) ([]*dto.GRPCRoom, error)
	Get(hotelID, roomID int64) (*dto.GRPCRoom, error)
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
	panic("implement me")
}

func (r *serverAPI) Get(ctx context.Context, request *roomsv1.GetRequest) (*roomsv1.GetResponse, error) {
	//TODO implement me
	panic("implement me")
}

func validateGetAll(r *roomsv1.GetAllRequest) error {
	if r.HotelID
}

func validateGet(r *roomsv1.GetAllRequest) error {

}
