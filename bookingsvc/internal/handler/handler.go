package bookingv1

import (
	"context"
	"errors"
	"time"

	gen "github.com/olegetoya/booking/bookingsvc/internal/gen/bookingv1"
)

type BookingService interface {
	GetAvailableRooms(
		ctx context.Context,
		hotelID int64,
		dateFrom time.Time,
		dateTo time.Time,
	) ([]RoomDTO, error)

	CreateBooking(
		ctx context.Context,
		userID int64,
		hotelID int64,
		roomID int64,
		dateFrom time.Time,
		dateTo time.Time,
	) (BookingDTO, error)

	GetBookingByID(
		ctx context.Context,
		bookingID int64,
	) (BookingDTO, error)

	CancelBooking(
		ctx context.Context,
		bookingID int64,
	) error
}

type Handler struct {
	service BookingService
}

func NewHandler(service BookingService) *Handler {
	return &Handler{
		service: service,
	}
}

type RoomDTO struct {
	ID          int64
	HotelID     int64
	RoomNum     int64
	Type        string
	Cost        int64
	IsAvailable bool
}

type BookingDTO struct {
	ID       int64
	UserID   int64
	HotelID  int64
	RoomID   int64
	DateFrom time.Time
	DateTo   time.Time
	Status   string
}

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyBooked = errors.New("already booked")
)

func (h *Handler) GetAvailableRooms(
	ctx context.Context,
	params gen.GetAvailableRoomsParams,
) (gen.GetAvailableRoomsRes, error) {
	dateFrom, err := time.Parse("2006-01-02", params.DateFrom)
	if err != nil {
		return &gen.ErrorResponse{
			Error: "invalid date_from",
		}, nil
	}

	dateTo, err := time.Parse("2006-01-02", params.DateTo)
	if err != nil {
		return &gen.ErrorResponse{
			Error: "invalid date_to",
		}, nil
	}

	rooms, err := h.service.GetAvailableRooms(
		ctx,
		params.HotelID,
		dateFrom,
		dateTo,
	)
	if err != nil {
		return &gen.ErrorResponse{
			Error: err.Error(),
		}, nil
	}

	return &gen.AvailableRoomsResponse{
		Rooms: mapRoomsToGen(rooms),
	}, nil
}

func (h *Handler) CreateBooking(
	ctx context.Context,
	req *gen.CreateBookingRequest,
) (gen.CreateBookingRes, error) {
	dateFrom, err := time.Parse("2006-01-02", req.DateFrom)
	if err != nil {
		return &gen.ErrorResponse{
			Error: "invalid date_from",
		}, nil
	}

	dateTo, err := time.Parse("2006-01-02", req.DateTo)
	if err != nil {
		return &gen.ErrorResponse{
			Error: "invalid date_to",
		}, nil
	}

	booking, err := h.service.CreateBooking(
		ctx,
		req.UserID,
		req.HotelID,
		req.RoomID,
		dateFrom,
		dateTo,
	)
	if err != nil {
		if errors.Is(err, ErrAlreadyBooked) {
			return &gen.ErrorResponse{
				Error: "room already booked for selected dates",
			}, nil
		}

		return &gen.ErrorResponse{
			Error: err.Error(),
		}, nil
	}

	return &gen.BookingResponse{
		ID:       booking.ID,
		UserID:   booking.UserID,
		HotelID:  booking.HotelID,
		RoomID:   booking.RoomID,
		DateFrom: booking.DateFrom.Format("2006-01-02"),
		DateTo:   booking.DateTo.Format("2006-01-02"),
		Status:   booking.Status,
	}, nil
}

func (h *Handler) GetBookingByID(
	ctx context.Context,
	params gen.GetBookingByIDParams,
) (gen.GetBookingByIDRes, error) {
	booking, err := h.service.GetBookingByID(ctx, params.BookingID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return &gen.ErrorResponse{
				Error: "booking not found",
			}, nil
		}

		return &gen.ErrorResponse{
			Error: err.Error(),
		}, nil
	}

	return &gen.BookingResponse{
		ID:       booking.ID,
		UserID:   booking.UserID,
		HotelID:  booking.HotelID,
		RoomID:   booking.RoomID,
		DateFrom: booking.DateFrom.Format("2006-01-02"),
		DateTo:   booking.DateTo.Format("2006-01-02"),
		Status:   booking.Status,
	}, nil
}

func (h *Handler) CancelBooking(
	ctx context.Context,
	params gen.CancelBookingParams,
) (gen.CancelBookingRes, error) {
	err := h.service.CancelBooking(ctx, params.BookingID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return &gen.ErrorResponse{
				Error: "booking not found",
			}, nil
		}

		return &gen.ErrorResponse{
			Error: err.Error(),
		}, nil
	}

	return &gen.CancelBookingNoContent{}, nil
}

func mapRoomsToGen(rooms []RoomDTO) []gen.Room {
	result := make([]gen.Room, 0, len(rooms))

	for _, room := range rooms {
		result = append(result, gen.Room{
			ID:          room.ID,
			HotelID:     room.HotelID,
			RoomNum:     room.RoomNum,
			Type:        room.Type,
			Cost:        room.Cost,
			IsAvailable: room.IsAvailable,
		})
	}

	return result
}
