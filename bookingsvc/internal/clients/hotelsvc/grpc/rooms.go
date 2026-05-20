package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/olegetoya/booking/bookingsvc/internal/domain"
	roomsv1 "github.com/olegetoya/booking/protos/gen/go/hotelsvc/rooms"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RoomsClient struct {
	api     roomsv1.RoomsClient
	timeout time.Duration
}

func NewRoomsClient(addr string, timeout time.Duration) (*RoomsClient, error) {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("create hotelsvc grpc client: %w", err)
	}

	return &RoomsClient{
		api:     roomsv1.NewRoomsClient(conn),
		timeout: timeout,
	}, nil
}

func (c *RoomsClient) GetAllRooms(
	ctx context.Context,
	hotelID int64,
) ([]domain.Room, error) {
	const op = "clients.hotelsvc.grpc.RoomsClient.GetAllRooms"

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.api.GetAll(ctx, &roomsv1.GetAllRequest{
		HotelID: hotelID,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: get all rooms: %w", op, err)
	}

	rooms := make([]domain.Room, 0, len(resp.Rooms))

	for _, room := range resp.Rooms {
		rooms = append(rooms, mapRoomFromProto(room))
	}

	return rooms, nil
}

func (c *RoomsClient) GetRoom(
	ctx context.Context,
	hotelID int64,
	roomID int64,
) (domain.Room, error) {
	const op = "clients.hotelsvc.grpc.RoomsClient.GetRoom"

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.api.Get(ctx, &roomsv1.GetRequest{
		HotelID: hotelID,
		RoomID:  roomID,
	})
	if err != nil {
		return domain.Room{}, fmt.Errorf("%s: get room: %w", op, err)
	}

	return mapRoomFromProto(resp.Room), nil
}

func mapRoomFromProto(room *roomsv1.Room) domain.Room {
	if room == nil {
		return domain.Room{}
	}

	return domain.Room{
		ID:          room.RoomID,
		HotelID:     room.HotelID,
		RoomNum:     int64(room.RoomNum),
		Type:        room.Type,
		Cost:        room.Cost,
		IsAvailable: room.IsAvailable != 0,
	}
}
