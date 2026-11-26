package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/olegetoya/booking/hotelsvc/internal/converter"
	"github.com/olegetoya/booking/hotelsvc/internal/dto"
	"github.com/olegetoya/booking/hotelsvc/internal/model"
	"github.com/olegetoya/booking/hotelsvc/internal/my_errors"
)

type RoomService struct {
	repo RoomStorage
}

func NewRoomService(repo RoomStorage) *RoomService {
	return &RoomService{
		repo: repo,
	}
}

type RoomStorage interface {
	Create(ctx context.Context, hotelId int64, room *model.Room) (int64, error)
	Update(ctx context.Context, hotelId, id int64, room *model.Room) error
	GetAll(ctx context.Context, hotelId int64) ([]*model.Room, error)
	Get(ctx context.Context, hotelId, id int64) (*model.Room, error)
	Delete(ctx context.Context, hotelId, id int64) error
}

func (r *RoomService) Create(ctx context.Context, hotelId int64, room *dto.Room) error {
	err := validateRoomInfo(room)
	if err != nil {
		return fmt.Errorf("service create room: %w", errors.Join(my_errors.RoomValidationErr, err))
	}
	roomConv, err := converter.ConvertFromRoomDTO(room)
	if err != nil {
		return fmt.Errorf("service create room: convert dto to model: %w", err)
	}
	_, err = r.repo.Create(ctx, hotelId, &roomConv)
	if err != nil {
		return fmt.Errorf("service create room: %w", err)
	}
	return nil
}

func (r *RoomService) Update(ctx context.Context, hotelId, id int64, room *dto.Room) error {
	err := validateRoomInfo(room)
	if err != nil {
		return fmt.Errorf("service update room id=%d: %w", id, errors.Join(my_errors.RoomValidationErr, err))
	}
	roomConv, err := converter.ConvertFromRoomDTO(room)
	if err != nil {
		return fmt.Errorf("service update room id=%d: convert dto to model: %w", id, err)
	}
	err = r.repo.Update(ctx, hotelId, id, &roomConv)
	if err != nil {
		return fmt.Errorf("service update room id=%d: %w", id, err)
	}
	return nil
}

func (r *RoomService) GetAll(ctx context.Context, hotelId int64) ([]*dto.Room, error) {
	rooms, err := r.repo.GetAll(ctx, hotelId)
	if err != nil {
		return nil, fmt.Errorf("service get all rooms hotelID=%d: %w", hotelId, err)
	}

	roomsDTO := make([]*dto.Room, 0, len(rooms))
	for _, room := range rooms {
		roomDTO, err := converter.ConvertToRoomDTO(room)
		if err != nil {
			return nil, fmt.Errorf("service get all rooms: convert model to dto: %w", err)
		}
		roomsDTO = append(roomsDTO, &roomDTO)
	}

	return roomsDTO, nil
}

func (r *RoomService) Get(ctx context.Context, hotelId, id int64) (*dto.Room, error) {
	room, err := r.repo.Get(ctx, hotelId, id)
	if err != nil {
		return nil, fmt.Errorf("service get room id=%d: %w", id, err)
	}

	roomDTO, err := converter.ConvertToRoomDTO(room)
	if err != nil {
		return nil, fmt.Errorf("service get room id=%d: convert model to dto: %w", id, err)
	}

	return &roomDTO, nil
}

func (r *RoomService) Delete(ctx context.Context, hotelId, id int64) error {
	err := r.repo.Delete(ctx, hotelId, id)
	if err != nil {
		return fmt.Errorf("service delete room id=%d: %w", id, err)
	}
	return nil
}

func isValidRoomType(r string) bool {
	switch r {
	case "economic", "standard", "luxury":
		return true
	default:
		return false
	}
}

func isValidRoomAvailability(r int) bool {
	switch r {
	case 0, 1, 2:
		return true
	default:
		return false
	}
}

func validateRoomInfo(room *dto.Room) error {
	if room == nil {
		return my_errors.NilPointerToRoomErr
	}
	if !isValidRoomType(room.Type) {
		return my_errors.InvalidRoomTypeErr
	}
	if !isValidRoomAvailability(room.IsAvailable) {
		return my_errors.InvalidAvailabilityErr
	}
	if room.RoomNum < 1 {
		return my_errors.InvalidRoomNumberErr
	}
	if room.Cost < 0 {
		return my_errors.InvalidRoomCostErr
	}
	return nil
}
