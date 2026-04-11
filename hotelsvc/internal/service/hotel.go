package service

import (
	"errors"
	"fmt"

	"github.com/olegetoya/booking/hotelsvc/internal/converter"
	"github.com/olegetoya/booking/hotelsvc/internal/dto"
	"github.com/olegetoya/booking/hotelsvc/internal/model"
	"github.com/olegetoya/booking/hotelsvc/internal/my_errors"
)

type HotelService struct {
	repo HotelStorage
}

func NewHotelService(repo HotelStorage) *HotelService {
	return &HotelService{
		repo: repo,
	}
}

type HotelStorage interface {
	Create(hotel *model.Hotel) (int64, error)
	Update(id int64, hotel *model.Hotel) error
	GetAll() ([]*model.Hotel, error)
	Get(id int64) (*model.Hotel, error)
	Delete(id int64) error
}

func (h *HotelService) Create(hotelDTO *dto.Hotel) error {
	if err := validateHotelInfo(hotelDTO); err != nil {
		return fmt.Errorf(
			"service create hotel: %w",
			errors.Join(my_errors.HotelValidationErr, err),
		)
	}

	hotel, err := converter.ConvertFromHotelDTO(hotelDTO)
	if err != nil {
		return fmt.Errorf("service create hotel: convert dto to model: %w", err)
	}

	_, err = h.repo.Create(&hotel)
	if err != nil {
		return fmt.Errorf("service create hotel: %w", err)
	}

	return nil
}

func (h *HotelService) Update(id int64, hotelDTO *dto.Hotel) error {
	if err := validateHotelInfo(hotelDTO); err != nil {
		return fmt.Errorf(
			"service update hotel id=%d: %w",
			id,
			errors.Join(my_errors.HotelValidationErr, err),
		)
	}

	hotel, err := converter.ConvertFromHotelDTO(hotelDTO)
	if err != nil {
		return fmt.Errorf("service update hotel id=%d: convert dto to model: %w", id, err)
	}

	if err = h.repo.Update(id, &hotel); err != nil {
		return fmt.Errorf("service update hotel id=%d: %w", id, err)
	}

	return nil
}

func (h *HotelService) GetAll() ([]*dto.Hotel, error) {
	hotels, err := h.repo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("service get all hotels: %w", err)
	}

	hotelsDTO := make([]*dto.Hotel, 0, len(hotels))
	for _, hotel := range hotels {
		hotelDTO, err := converter.ConvertToHotelDTO(hotel)
		if err != nil {
			return nil, fmt.Errorf("service get all hotels: convert model to dto: %w", err)
		}
		hotelsDTO = append(hotelsDTO, &hotelDTO)
	}

	return hotelsDTO, nil
}

func (h *HotelService) Get(id int64) (*dto.Hotel, error) {
	hotel, err := h.repo.Get(id)
	if err != nil {
		return nil, fmt.Errorf("service get hotel id=%d: %w", id, err)
	}

	hotelDTO, err := converter.ConvertToHotelDTO(hotel)
	if err != nil {
		return nil, fmt.Errorf("service get hotel id=%d: convert model to dto: %w", id, err)
	}

	return &hotelDTO, nil
}

func (h *HotelService) Delete(id int64) error {
	if err := h.repo.Delete(id); err != nil {
		return fmt.Errorf("service delete hotel id=%d: %w", id, err)
	}
	return nil
}

func isValidHotelRating(r int) bool {
	switch r {
	case 0, 3, 4, 5:
		return true
	default:
		return false
	}
}

func validateHotelInfo(hotel *dto.Hotel) error {
	if hotel == nil {
		return my_errors.NilPointerToHotelErr
	}
	if hotel.Name == "" {
		return my_errors.HotelWithoutNameErr
	}
	if !isValidHotelRating(hotel.Rating) {
		return my_errors.InvalidHotelRatingErr
	}
	if hotel.RoomsNum < 0 {
		return my_errors.InvalidRoomsNumberErr
	}
	if hotel.RoomsOccupied < 0 || hotel.RoomsOccupied > hotel.RoomsNum {
		return my_errors.InvalidRoomsOccupiedNumberErr
	}
	return nil
}
