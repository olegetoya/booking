package grpcSrv

import (
	"context"
	"fmt"
	"github.com/olegetoya/booking/hotelsvc/internal/converter"
	"github.com/olegetoya/booking/hotelsvc/internal/model"
	roomsv1 "github.com/olegetoya/booking/protos/gen/go/hotelsvc/rooms"
	"log/slog"
)

type RoomStorage interface {
	GetAll(hotelId int64) ([]*model.Room, error)
	Get(hotelId, id int64) (*model.Room, error)
}

type RoomsServer struct {
	repo RoomStorage
}

func NewRoomsServer(repo RoomStorage) *RoomsServer {
	return &RoomsServer{
		repo: repo,
	}
}

func (srv *RoomsServer) GetAll(ctx context.Context, hotelId int64) ([]*roomsv1.Room, error) {
	rooms, err := srv.repo.GetAll(hotelId)
	if err != nil {
		slog.Warn("err trying to get all rooms: %v", slog.String("err", err.Error()))
		return nil, fmt.Errorf("service get all rooms hotelID=%d: %w", hotelId, err)
	}

	roomsGRPC := make([]*roomsv1.Room, 0, len(rooms))
	for _, room := range rooms {
		roomGRPC, err := converter.ConvertToGRPCRoom(room)
		if err != nil {
			return nil, fmt.Errorf("service get all rooms: convert model to grpc model: %w", err)
		}
		roomsGRPC = append(roomsGRPC, roomGRPC)
	}

	return roomsGRPC, nil
}

func (srv *RoomsServer) Get(ctx context.Context, hotelId, id int64) (*roomsv1.Room, error) {
	room, err := srv.repo.Get(hotelId, id)
	if err != nil {
		slog.Warn("err trying to get room: %v", slog.String("err", err.Error()))
		return nil, fmt.Errorf("service get room id=%d: %w", id, err)
	}
	roomGRPC, err := converter.ConvertToGRPCRoom(room)
	if err != nil {
		return nil, fmt.Errorf("service get room id=%d: convert model to grpc model: %w", id, err)
	}

	return roomGRPC, nil
}
