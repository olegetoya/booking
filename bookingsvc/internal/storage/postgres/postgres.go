package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/olegetoya/booking/bookingsvc/internal/domain"
)

type BookingRepository struct {
	db *sql.DB
}

func NewBookingRepository(db *sql.DB) *BookingRepository {
	return &BookingRepository{
		db: db,
	}
}

func (r *BookingRepository) GetBookedRoomIDs(
	ctx context.Context,
	hotelID int64,
	dateFrom time.Time,
	dateTo time.Time,
) ([]int64, error) {
	const op = "storage.postgres.BookingRepository.GetBookedRoomIDs"

	query := `
		SELECT room_id
		FROM bookings
		WHERE hotel_id = $1
		  AND status = 'active'
		  AND date_from < $3
		  AND date_to > $2
	`

	rows, err := r.db.QueryContext(ctx, query, hotelID, dateFrom, dateTo)
	if err != nil {
		return nil, fmt.Errorf("%s: query booked room ids: %w", op, err)
	}
	defer rows.Close()

	roomIDs := make([]int64, 0)

	for rows.Next() {
		var roomID int64

		if err := rows.Scan(&roomID); err != nil {
			return nil, fmt.Errorf("%s: scan room id: %w", op, err)
		}

		roomIDs = append(roomIDs, roomID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows error: %w", op, err)
	}

	return roomIDs, nil
}

func (r *BookingRepository) CreateBooking(
	ctx context.Context,
	booking domain.Booking,
) (domain.Booking, error) {
	const op = "storage.postgres.BookingRepository.CreateBooking"

	query := `
    INSERT INTO bookings (
        user_id,
        hotel_id,
        room_id,
        date_from,
        date_to,
        price_per_night,
        total_cost,
        status
    )
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    RETURNING id
`

	var created domain.Booking

	err := r.db.QueryRowContext(
		ctx,
		query,
		booking.UserID,
		booking.HotelID,
		booking.RoomID,
		booking.DateFrom,
		booking.DateTo,
		booking.PricePerNight,
		booking.TotalCost,
		booking.Status,
	).Scan(&booking.ID)
	if err != nil {
		return domain.Booking{}, fmt.Errorf("%s: insert booking: %w", op, err)
	}

	return created, nil
}

func (r *BookingRepository) GetBookings(
	ctx context.Context,
	userID *int64,
	hotelID *int64,
) ([]domain.Booking, error) {
	const op = "storage.postgres.BookingRepository.GetBookings"

	query := `
		SELECT
			id,
			user_id,
			hotel_id,
			room_id,
			date_from,
			date_to,
			price_per_night,
    		total_cost,
			status
		FROM bookings
		WHERE ($1::bigint IS NULL OR user_id = $1)
		  AND ($2::bigint IS NULL OR hotel_id = $2)
		ORDER BY id DESC
	`

	rows, err := r.db.QueryContext(
		ctx,
		query,
		userID,
		hotelID,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: query bookings: %w", op, err)
	}
	defer rows.Close()

	bookings := make([]domain.Booking, 0)

	for rows.Next() {
		var booking domain.Booking

		if err := rows.Scan(
			&booking.ID,
			&booking.UserID,
			&booking.HotelID,
			&booking.RoomID,
			&booking.DateFrom,
			&booking.DateTo,
			&booking.PricePerNight,
			&booking.TotalCost,
			&booking.Status,
		); err != nil {
			return nil, fmt.Errorf("%s: scan booking: %w", op, err)
		}

		bookings = append(bookings, booking)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows error: %w", op, err)
	}

	return bookings, nil
}

func (r *BookingRepository) GetBookingByID(
	ctx context.Context,
	bookingID int64,
) (domain.Booking, error) {
	const op = "storage.postgres.BookingRepository.GetBookingByID"

	query := `
		SELECT id, user_id, hotel_id, room_id, date_from, date_to, price_per_night, total_cost, status
		FROM bookings
		WHERE id = $1
	`

	var booking domain.Booking

	err := r.db.QueryRowContext(ctx, query, bookingID).Scan(
		&booking.ID,
		&booking.UserID,
		&booking.HotelID,
		&booking.RoomID,
		&booking.DateFrom,
		&booking.DateTo,
		&booking.PricePerNight,
		&booking.TotalCost,
		&booking.Status,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Booking{}, fmt.Errorf("%s: %w", op, domain.ErrBookingNotFound)
		}

		return domain.Booking{}, fmt.Errorf("%s: select booking by id: %w", op, err)
	}

	return booking, nil
}

func (r *BookingRepository) CancelBooking(
	ctx context.Context,
	bookingID int64,
) error {
	const op = "storage.postgres.BookingRepository.CancelBooking"

	query := `
		UPDATE bookings
		SET status = 'cancelled'
		WHERE id = $1
		  AND status = 'active'
	`

	result, err := r.db.ExecContext(ctx, query, bookingID)
	if err != nil {
		return fmt.Errorf("%s: cancel booking: %w", op, err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: get affected rows: %w", op, err)
	}

	if affected == 0 {
		return fmt.Errorf("%s: %w", op, domain.ErrBookingNotFound)
	}

	return nil
}
