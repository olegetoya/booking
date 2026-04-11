package converter

import (
	"time"

	"github.com/olegetoya/booking/hotelsvc/internal/dto"
	"github.com/olegetoya/booking/hotelsvc/internal/model"
)

func ConvertFromHotelDTO(d *dto.Hotel) (model.Hotel, error) {
	var hotel model.Hotel

	hotel.Id = 0
	hotel.Rating = model.HotelRating(d.Rating)
	hotel.Name = d.Name
	hotel.Address = d.Address
	hotel.RoomsNum = d.RoomsNum
	hotel.RoomsOccupied = d.RoomsOccupied
	hotel.CreatedAt = time.Time{}
	hotel.UpdatedAt = time.Time{}

	return hotel, nil
}

func ConvertToHotelDTO(h *model.Hotel) (dto.Hotel, error) {
	var hotel dto.Hotel

	hotel.Rating = int(h.Rating)
	hotel.Name = h.Name
	hotel.Address = h.Address
	hotel.RoomsNum = h.RoomsNum
	hotel.RoomsOccupied = h.RoomsOccupied

	return hotel, nil
}

func ConvertFromRoomDTO(d *dto.Room) (model.Room, error) {
	var room model.Room

	room.Id = 0
	room.HotelId = 0
	room.RoomNum = d.RoomNum
	room.Type = model.RoomType(d.Type)
	room.Cost = d.Cost
	room.IsAvailable = model.Availability(d.IsAvailable)
	room.CreatedAt = time.Time{}
	room.UpdatedAt = time.Time{}

	return room, nil
}

func ConvertToRoomDTO(r *model.Room) (dto.Room, error) {
	var room dto.Room

	room.RoomNum = r.RoomNum
	room.Type = string(r.Type)
	room.Cost = r.Cost
	room.IsAvailable = int(r.IsAvailable)

	return room, nil
}
