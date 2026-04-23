package grpcSrv

import (
	"github.com/olegetoya/booking/hotelsvc/internal/model"
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
