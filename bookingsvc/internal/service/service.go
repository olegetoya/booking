package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/olegetoya/booking/bookingsvc/internal/domain"
	"github.com/olegetoya/booking/bookingsvc/internal/events"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
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

	GetBookings(
		ctx context.Context,
		userID *int64,
		hotelID *int64,
	) ([]domain.Booking, error)

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

type EventPublisher interface {
	Produce(
		ctx context.Context,
		key string,
		value any,
	) error
}

type BookingService struct {
	log         *slog.Logger
	repo        BookingRepository
	roomsClient RoomsClient
	publisher   EventPublisher
}

func NewBookingService(
	log *slog.Logger,
	repo BookingRepository,
	roomsClient RoomsClient,
	eventPublisher EventPublisher,
) *BookingService {
	return &BookingService{
		log:         log,
		repo:        repo,
		roomsClient: roomsClient,
		publisher:   eventPublisher,
	}
}

func (s *BookingService) GetAvailableRooms(
	ctx context.Context,
	hotelID int64,
	dateFrom time.Time,
	dateTo time.Time,
) (_ []domain.AvailableRoom, resultErr error) {
	const op = "service.BookingService.GetAvailableRooms"
	ctx, span := otel.Tracer("bookingsvc/service").Start(ctx, "BookingService.GetAvailableRooms", trace.WithAttributes(attribute.Int64("hotel.id", hotelID)))
	defer func() { finishBookingSpan(span, resultErr) }()

	log := s.log.With(
		slog.String("op", op),
		slog.Int64("hotel_id", hotelID),
		slog.Time("date_from", dateFrom),
		slog.Time("date_to", dateTo),
	)

	if !validDateRange(dateFrom, dateTo) {
		log.WarnContext(ctx, "invalid date range")
		return nil, fmt.Errorf("%s: %w", op, domain.ErrInvalidDateRange)
	}

	nights := int64(dateTo.Sub(dateFrom) / (24 * time.Hour))

	rooms, err := s.roomsClient.GetAllRooms(ctx, hotelID)
	if err != nil {
		log.ErrorContext(ctx,
			"failed to get rooms from hotelsvc",
			slog.Any("error", err),
		)

		return nil, fmt.Errorf(
			"%s: get all rooms from hotelsvc: %w",
			op,
			err,
		)
	}

	bookedRoomIDs, err := s.repo.GetBookedRoomIDs(
		ctx,
		hotelID,
		dateFrom,
		dateTo,
	)
	if err != nil {
		log.ErrorContext(ctx,
			"failed to get booked room ids",
			slog.Any("error", err),
		)

		return nil, fmt.Errorf(
			"%s: get booked room ids: %w",
			op,
			err,
		)
	}

	booked := make(map[int64]struct{}, len(bookedRoomIDs))

	for _, roomID := range bookedRoomIDs {
		booked[roomID] = struct{}{}
	}

	availableRooms := make(
		[]domain.AvailableRoom,
		0,
		len(rooms),
	)

	for _, room := range rooms {
		if !room.IsAvailable {
			continue
		}

		if _, exists := booked[room.ID]; exists {
			continue
		}

		totalCost := room.Cost * nights

		availableRooms = append(
			availableRooms,
			domain.AvailableRoom{
				ID:           room.ID,
				HotelID:      room.HotelID,
				RoomNum:      room.RoomNum,
				Type:         room.Type,
				CostPerNight: room.Cost,
				TotalCost:    totalCost,
				IsAvailable:  room.IsAvailable,
			},
		)
	}

	log.InfoContext(ctx,
		"available rooms received",
		slog.Int("total_rooms", len(rooms)),
		slog.Int("booked_rooms", len(bookedRoomIDs)),
		slog.Int("available_rooms", len(availableRooms)),
		slog.Int64("nights", nights),
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
) (_ domain.Booking, resultErr error) {
	const op = "service.BookingService.CreateBooking"
	ctx, span := otel.Tracer("bookingsvc/service").Start(ctx, "BookingService.CreateBooking", trace.WithAttributes(attribute.Int64("hotel.id", hotelID), attribute.Int64("room.id", roomID)))
	defer func() { finishBookingSpan(span, resultErr) }()

	log := s.log.With(
		slog.String("op", op),
		slog.Int64("user_id", userID),
		slog.Int64("hotel_id", hotelID),
		slog.Int64("room_id", roomID),
		slog.Time("date_from", dateFrom),
		slog.Time("date_to", dateTo),
	)

	if !validDateRange(dateFrom, dateTo) {
		log.WarnContext(ctx, "invalid date range")
		return domain.Booking{}, fmt.Errorf("%s: %w", op, domain.ErrInvalidDateRange)
	}

	room, err := s.roomsClient.GetRoom(ctx, hotelID, roomID)
	if err != nil {
		if errors.Is(err, domain.ErrRoomNotFound) {
			log.WarnContext(ctx, "room not found")
			return domain.Booking{}, fmt.Errorf("%s: %w", op, domain.ErrRoomNotFound)
		}

		log.ErrorContext(ctx,
			"failed to get room from hotelsvc",
			slog.Any("error", err),
		)

		return domain.Booking{},
			fmt.Errorf("%s: get room from hotelsvc: %w", op, err)
	}

	if room.HotelID != hotelID {
		log.WarnContext(ctx,
			"room belongs to another hotel",
			slog.Int64("actual_hotel_id", room.HotelID),
		)

		return domain.Booking{},
			fmt.Errorf("%s: %w", op, domain.ErrRoomFromOtherHotel)
	}

	if !room.IsAvailable {
		log.WarnContext(ctx, "room is marked as unavailable")

		return domain.Booking{},
			fmt.Errorf("%s: %w", op, domain.ErrRoomAlreadyBooked)
	}

	bookedRoomIDs, err := s.repo.GetBookedRoomIDs(
		ctx,
		hotelID,
		dateFrom,
		dateTo,
	)
	if err != nil {
		log.ErrorContext(ctx,
			"failed to check booked rooms",
			slog.Any("error", err),
		)

		return domain.Booking{},
			fmt.Errorf("%s: check booked rooms: %w", op, err)
	}

	for _, bookedRoomID := range bookedRoomIDs {
		if bookedRoomID == roomID {
			log.WarnContext(ctx, "room already booked for selected dates")

			return domain.Booking{},
				fmt.Errorf("%s: %w", op, domain.ErrRoomAlreadyBooked)
		}
	}

	nights := int64(dateTo.Sub(dateFrom) / (24 * time.Hour))

	pricePerNight := room.Cost
	totalCost := pricePerNight * nights

	booking := domain.Booking{
		UserID:        userID,
		HotelID:       hotelID,
		RoomID:        roomID,
		DateFrom:      dateFrom,
		DateTo:        dateTo,
		PricePerNight: pricePerNight,
		TotalCost:     totalCost,
		Status:        "active",
	}

	createdBooking, err := s.repo.CreateBooking(ctx, booking)
	if err != nil {
		log.ErrorContext(ctx,
			"failed to create booking",
			slog.Any("error", err),
		)

		return domain.Booking{},
			fmt.Errorf("%s: create booking in repository: %w", op, err)
	}

	span.SetAttributes(attribute.Int64("booking.id", createdBooking.ID))

	event := events.BookingCreated{
		EventID:    uuid.NewString(),
		EventType:  "booking.created",
		BookingID:  createdBooking.ID,
		UserID:     createdBooking.UserID,
		HotelID:    createdBooking.HotelID,
		RoomID:     createdBooking.RoomID,
		DateFrom:   createdBooking.DateFrom,
		DateTo:     createdBooking.DateTo,
		TotalCost:  createdBooking.TotalCost,
		OccurredAt: time.Now(),
	}

	if err := s.publisher.Produce(
		ctx,
		strconv.FormatInt(createdBooking.ID, 10),
		event,
	); err != nil {
		return domain.Booking{}, fmt.Errorf("%s: produce booking created event: %w", op, err)
	}

	log.InfoContext(ctx,
		"booking created",
		slog.Int64("booking_id", createdBooking.ID),
		slog.Int64("price_per_night", createdBooking.PricePerNight),
		slog.Int64("total_cost", createdBooking.TotalCost),
		slog.Int64("nights", nights),
	)

	return createdBooking, nil
}

func (s *BookingService) GetBookings(
	ctx context.Context,
	userID *int64,
	hotelID *int64,
) (_ []domain.Booking, resultErr error) {
	const op = "service.BookingService.GetBookings"
	ctx, span := otel.Tracer("bookingsvc/service").Start(ctx, "BookingService.GetBookings")
	defer func() { finishBookingSpan(span, resultErr) }()

	log := s.log.With(
		slog.String("op", op),
	)

	if userID != nil {
		log = log.With(slog.Int64("user_id", *userID))
	}

	if hotelID != nil {
		span.SetAttributes(attribute.Int64("hotel.id", *hotelID))
		log = log.With(slog.Int64("hotel_id", *hotelID))
	}

	bookings, err := s.repo.GetBookings(ctx, userID, hotelID)
	if err != nil {
		log.ErrorContext(ctx,
			"failed to get bookings",
			slog.Any("error", err),
		)

		return nil, fmt.Errorf("%s: get bookings: %w", op, err)
	}

	log.InfoContext(ctx,
		"bookings received",
		slog.Int("count", len(bookings)),
	)

	return bookings, nil
}

func (s *BookingService) GetBookingByID(
	ctx context.Context,
	bookingID int64,
) (_ domain.Booking, resultErr error) {
	const op = "service.BookingService.GetBookingByID"
	ctx, span := otel.Tracer("bookingsvc/service").Start(ctx, "BookingService.GetBookingByID", trace.WithAttributes(attribute.Int64("booking.id", bookingID)))
	defer func() { finishBookingSpan(span, resultErr) }()

	log := s.log.With(
		slog.String("op", op),
		slog.Int64("booking_id", bookingID),
	)

	booking, err := s.repo.GetBookingByID(ctx, bookingID)
	if err != nil {
		if errors.Is(err, domain.ErrBookingNotFound) {
			log.WarnContext(ctx, "booking not found")
			return domain.Booking{}, fmt.Errorf("%s: %w", op, domain.ErrBookingNotFound)
		}

		log.ErrorContext(ctx, "failed to get booking by id", slog.Any("error", err))
		return domain.Booking{}, fmt.Errorf("%s: get booking by id: %w", op, err)
	}

	return booking, nil
}

func (s *BookingService) CancelBooking(
	ctx context.Context,
	bookingID int64,
) (resultErr error) {
	const op = "service.BookingService.CancelBooking"
	ctx, span := otel.Tracer("bookingsvc/service").Start(ctx, "BookingService.CancelBooking", trace.WithAttributes(attribute.Int64("booking.id", bookingID)))
	defer func() { finishBookingSpan(span, resultErr) }()

	log := s.log.With(
		slog.String("op", op),
		slog.Int64("booking_id", bookingID),
	)

	err := s.repo.CancelBooking(ctx, bookingID)
	if err != nil {
		if errors.Is(err, domain.ErrBookingNotFound) {
			log.WarnContext(ctx, "booking not found")
			return fmt.Errorf("%s: %w", op, domain.ErrBookingNotFound)
		}

		log.ErrorContext(ctx, "failed to cancel booking", slog.Any("error", err))
		return fmt.Errorf("%s: cancel booking: %w", op, err)
	}

	event := events.BookingCancelled{
		EventID:    uuid.NewString(),
		EventType:  "booking.cancelled",
		BookingID:  bookingID,
		OccurredAt: time.Now(),
	}

	if err := s.publisher.Produce(
		ctx,
		strconv.FormatInt(bookingID, 10),
		event,
	); err != nil {
		return fmt.Errorf("%s: produce booking cancelled event: %w", op, err)
	}

	log.InfoContext(ctx, "booking cancelled")

	return nil
}

func validDateRange(dateFrom time.Time, dateTo time.Time) bool {
	return dateTo.After(dateFrom)
}

func finishBookingSpan(span trace.Span, err error) {
	defer span.End()
	switch {
	case err == nil:
	case errors.Is(err, domain.ErrInvalidDateRange):
		span.SetAttributes(attribute.String("booking.outcome", "invalid_date_range"))
	case errors.Is(err, domain.ErrRoomAlreadyBooked):
		span.SetAttributes(attribute.String("booking.outcome", "room_unavailable"))
	case errors.Is(err, domain.ErrRoomNotFound):
		span.SetAttributes(attribute.String("booking.outcome", "room_not_found"))
	case errors.Is(err, domain.ErrRoomFromOtherHotel):
		span.SetAttributes(attribute.String("booking.outcome", "hotel_mismatch"))
	case errors.Is(err, domain.ErrBookingNotFound):
		span.SetAttributes(attribute.String("booking.outcome", "booking_not_found"))
	default:
		span.RecordError(errors.New("booking operation failed"))
		span.SetStatus(codes.Error, "booking operation failed")
	}
}
