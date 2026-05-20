package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/olegetoya/booking/bookingsvc/internal/domain"
)

type BookingRepository interface {
	GetBookedRoomIDs(
		ctx context.Context,
		hotelID int64,
		dateFrom time.Time,
		dateTo time.Time,
	) ([]int64, error)

	CreateBooking(
		ctx context.Context,
		booking domain.Booking,
	) (domain.Booking, error)

	GetBookingByID(
		ctx context.Context,
		bookingID int64,
	) (domain.Booking, error)

	CancelBooking(
		ctx context.Context,
		bookingID int64,
	) error
}

type RoomsClient interface {
	GetAllRooms(
		ctx context.Context,
		hotelID int64,
	) ([]domain.Room, error)

	GetRoom(
		ctx context.Context,
		hotelID int64,
		roomID int64,
	) (domain.Room, error)
}

type BookingService struct {
	log         *slog.Logger
	repo        BookingRepository
	roomsClient RoomsClient
}

func NewBookingService(
	log *slog.Logger,
	repo BookingRepository,
	roomsClient RoomsClient,
) *BookingService {
	return &BookingService{
		log:         log,
		repo:        repo,
		roomsClient: roomsClient,
	}
}

func (s *BookingService) GetAvailableRooms(
	ctx context.Context,
	hotelID int64,
	dateFrom time.Time,
	dateTo time.Time,
) ([]domain.Room, error) {
	const op = "service.BookingService.GetAvailableRooms"

	log := s.log.With(
		slog.String("op", op),
		slog.Int64("hotel_id", hotelID),
		slog.Time("date_from", dateFrom),
		slog.Time("date_to", dateTo),
	)

	if !validDateRange(dateFrom, dateTo) {
		log.Warn("invalid date range")
		return nil, fmt.Errorf("%s: %w", op, domain.ErrInvalidDateRange)
	}

	rooms, err := s.roomsClient.GetAllRooms(ctx, hotelID)
	if err != nil {
		log.Error("failed to get rooms from hotelsvc", slog.Any("error", err))
		return nil, fmt.Errorf("%s: get all rooms from hotelsvc: %w", op, err)
	}

	bookedRoomIDs, err := s.repo.GetBookedRoomIDs(ctx, hotelID, dateFrom, dateTo)
	if err != nil {
		log.Error("failed to get booked room ids", slog.Any("error", err))
		return nil, fmt.Errorf("%s: get booked room ids: %w", op, err)
	}

	booked := make(map[int64]struct{}, len(bookedRoomIDs))
	for _, roomID := range bookedRoomIDs {
		booked[roomID] = struct{}{}
	}

	availableRooms := make([]domain.Room, 0, len(rooms))

	for _, room := range rooms {
		if !room.IsAvailable {
			continue
		}

		if _, exists := booked[room.ID]; exists {
			continue
		}

		availableRooms = append(availableRooms, room)
	}

	log.Info(
		"available rooms received",
		slog.Int("total_rooms", len(rooms)),
		slog.Int("booked_rooms", len(bookedRoomIDs)),
		slog.Int("available_rooms", len(availableRooms)),
	)

	return availableRooms, nil
}

func (s *BookingService) CreateBooking(
	ctx context.Context,
	userID int64,
	hotelID int64,
	roomID int64,
	dateFrom time.Time,
	dateTo time.Time,
) (domain.Booking, error) {
	const op = "service.BookingService.CreateBooking"

	log := s.log.With(
		slog.String("op", op),
		slog.Int64("user_id", userID),
		slog.Int64("hotel_id", hotelID),
		slog.Int64("room_id", roomID),
		slog.Time("date_from", dateFrom),
		slog.Time("date_to", dateTo),
	)

	if !validDateRange(dateFrom, dateTo) {
		log.Warn("invalid date range")
		return domain.Booking{}, fmt.Errorf("%s: %w", op, domain.ErrInvalidDateRange)
	}

	room, err := s.roomsClient.GetRoom(ctx, hotelID, roomID)
	if err != nil {
		if errors.Is(err, domain.ErrRoomNotFound) {
			log.Warn("room not found")
			return domain.Booking{}, fmt.Errorf("%s: %w", op, domain.ErrRoomNotFound)
		}

		log.Error("failed to get room from hotelsvc", slog.Any("error", err))
		return domain.Booking{}, fmt.Errorf("%s: get room from hotelsvc: %w", op, err)
	}

	if room.HotelID != hotelID {
		log.Warn(
			"room belongs to another hotel",
			slog.Int64("actual_hotel_id", room.HotelID),
		)

		return domain.Booking{}, fmt.Errorf("%s: %w", op, domain.ErrRoomFromOtherHotel)
	}

	if !room.IsAvailable {
		log.Warn("room is marked as unavailable")
		return domain.Booking{}, fmt.Errorf("%s: %w", op, domain.ErrRoomAlreadyBooked)
	}

	bookedRoomIDs, err := s.repo.GetBookedRoomIDs(ctx, hotelID, dateFrom, dateTo)
	if err != nil {
		log.Error("failed to check booked rooms", slog.Any("error", err))
		return domain.Booking{}, fmt.Errorf("%s: check booked rooms: %w", op, err)
	}

	for _, bookedRoomID := range bookedRoomIDs {
		if bookedRoomID == roomID {
			log.Warn("room already booked for selected dates")
			return domain.Booking{}, fmt.Errorf("%s: %w", op, domain.ErrRoomAlreadyBooked)
		}
	}

	booking := domain.Booking{
		UserID:   userID,
		HotelID:  hotelID,
		RoomID:   roomID,
		DateFrom: dateFrom,
		DateTo:   dateTo,
		Status:   "active",
	}

	createdBooking, err := s.repo.CreateBooking(ctx, booking)
	if err != nil {
		log.Error("failed to create booking", slog.Any("error", err))
		return domain.Booking{}, fmt.Errorf("%s: create booking in repository: %w", op, err)
	}

	log.Info(
		"booking created",
		slog.Int64("booking_id", createdBooking.ID),
	)

	return createdBooking, nil
}

func (s *BookingService) GetBookingByID(
	ctx context.Context,
	bookingID int64,
) (domain.Booking, error) {
	const op = "service.BookingService.GetBookingByID"

	log := s.log.With(
		slog.String("op", op),
		slog.Int64("booking_id", bookingID),
	)

	booking, err := s.repo.GetBookingByID(ctx, bookingID)
	if err != nil {
		if errors.Is(err, domain.ErrBookingNotFound) {
			log.Warn("booking not found")
			return domain.Booking{}, fmt.Errorf("%s: %w", op, domain.ErrBookingNotFound)
		}

		log.Error("failed to get booking by id", slog.Any("error", err))
		return domain.Booking{}, fmt.Errorf("%s: get booking by id: %w", op, err)
	}

	return booking, nil
}

func (s *BookingService) CancelBooking(
	ctx context.Context,
	bookingID int64,
) error {
	const op = "service.BookingService.CancelBooking"

	log := s.log.With(
		slog.String("op", op),
		slog.Int64("booking_id", bookingID),
	)

	err := s.repo.CancelBooking(ctx, bookingID)
	if err != nil {
		if errors.Is(err, domain.ErrBookingNotFound) {
			log.Warn("booking not found")
			return fmt.Errorf("%s: %w", op, domain.ErrBookingNotFound)
		}

		log.Error("failed to cancel booking", slog.Any("error", err))
		return fmt.Errorf("%s: cancel booking: %w", op, err)
	}

	log.Info("booking cancelled")

	return nil
}

func validDateRange(dateFrom time.Time, dateTo time.Time) bool {
	return dateTo.After(dateFrom)
}
