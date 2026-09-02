package bookingv1

import (
	"context"
	"errors"
	"github.com/olegetoya/booking/bookingsvc/internal/domain"
	"time"

	gen "github.com/olegetoya/booking/bookingsvc/internal/gen/bookingv1"
)

type BookingService interface {
	GetAvailableRooms(
		ctx context.Context,
		hotelID int64,
		dateFrom time.Time,
		dateTo time.Time,
	) ([]domain.Room, error)

	CreateBooking(
		ctx context.Context,
		userID int64,
		hotelID int64,
		roomID int64,
		dateFrom time.Time,
		dateTo time.Time,
	) (domain.Booking, error)

	GetBookings(
		ctx context.Context,
		userID *int64,
		hotelID *int64,
	) ([]domain.Booking, error)

	GetBookingByID(ctx context.Context, bookingID int64) (domain.Booking, error)

	CancelBooking(ctx context.Context, bookingID int64) error
}

type Handler struct {
	service BookingService
}

func NewHandler(service BookingService) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) GetAvailableRooms(
	ctx context.Context,
	params gen.GetAvailableRoomsParams,
) (gen.GetAvailableRoomsRes, error) {

	dateFrom := params.DateFrom
	dateTo := params.DateTo

	rooms, err := h.service.GetAvailableRooms(
		ctx,
		params.HotelID,
		dateFrom,
		dateTo,
	)
	if err != nil {
		return &gen.GetAvailableRoomsInternalServerError{
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

	dateFrom := req.DateFrom
	dateTo := req.DateTo

	booking, err := h.service.CreateBooking(
		ctx,
		req.UserID,
		req.HotelID,
		req.RoomID,
		dateFrom,
		dateTo,
	)
	if err != nil {
		if errors.Is(err, domain.ErrRoomAlreadyBooked) {
			return &gen.CreateBookingConflict{
				Error: "room already booked for selected dates",
			}, nil
		}

		return &gen.CreateBookingInternalServerError{
			Error: err.Error(),
		}, nil
	}

	return &gen.BookingResponse{
		ID:       booking.ID,
		UserID:   booking.UserID,
		HotelID:  booking.HotelID,
		RoomID:   booking.RoomID,
		DateFrom: booking.DateFrom,
		DateTo:   booking.DateTo,
		Status:   booking.Status,
	}, nil
}

func (h *Handler) GetBookings(
	ctx context.Context,
	params gen.GetBookingsParams,
) (gen.GetBookingsRes, error) {

	var userID *int64
	var hotelID *int64

	if value, ok := params.UserID.Get(); ok {
		userID = &value
	}

	if value, ok := params.HotelID.Get(); ok {
		hotelID = &value
	}

	if userID == nil && hotelID == nil {
		return &gen.GetBookingsBadRequest{
			Error: "user_id or hotel_id is required",
		}, nil
	}

	bookings, err := h.service.GetBookings(
		ctx,
		userID,
		hotelID,
	)
	if err != nil {
		return &gen.GetBookingsInternalServerError{
			Error: err.Error(),
		}, nil
	}

	return &gen.BookingsResponse{
		Bookings: mapBookingsToGen(bookings),
	}, nil
}

func (h *Handler) GetBookingByID(
	ctx context.Context,
	params gen.GetBookingByIDParams,
) (gen.GetBookingByIDRes, error) {
	booking, err := h.service.GetBookingByID(ctx, params.BookingID)
	if err != nil {
		if errors.Is(err, domain.ErrBookingNotFound) {
			return &gen.GetBookingByIDNotFound{
				Error: "booking not found",
			}, nil
		}

		return &gen.GetBookingByIDInternalServerError{
			Error: err.Error(),
		}, nil
	}

	return &gen.BookingResponse{
		ID:       booking.ID,
		UserID:   booking.UserID,
		HotelID:  booking.HotelID,
		RoomID:   booking.RoomID,
		DateFrom: booking.DateFrom,
		DateTo:   booking.DateTo,
		Status:   booking.Status,
	}, nil
}

func (h *Handler) CancelBooking(
	ctx context.Context,
	params gen.CancelBookingParams,
) (gen.CancelBookingRes, error) {
	err := h.service.CancelBooking(ctx, params.BookingID)
	if err != nil {
		if errors.Is(err, domain.ErrBookingNotFound) {
			return &gen.CancelBookingNotFound{
				Error: "booking not found",
			}, nil
		}

		return &gen.CancelBookingInternalServerError{
			Error: err.Error(),
		}, nil
	}

	return &gen.CancelBookingNoContent{}, nil
}

func mapRoomsToGen(rooms []domain.Room) []gen.Room {
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

func mapBookingsToGen(bookings []domain.Booking) []gen.BookingResponse {
	result := make([]gen.BookingResponse, 0, len(bookings))

	for _, booking := range bookings {
		result = append(result, gen.BookingResponse{
			ID:       booking.ID,
			UserID:   booking.UserID,
			HotelID:  booking.HotelID,
			RoomID:   booking.RoomID,
			DateFrom: booking.DateFrom,
			DateTo:   booking.DateTo,
			Status:   booking.Status,
		})
	}

	return result
}
