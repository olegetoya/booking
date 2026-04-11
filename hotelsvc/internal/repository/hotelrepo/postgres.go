package hotelrepo

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/olegetoya/booking/hotelsvc/internal/model"
	"github.com/olegetoya/booking/hotelsvc/internal/my_errors"
)

type HotelPostgres struct {
	db *sql.DB
}

func NewHotelPostgres(db *sql.DB) *HotelPostgres {
	return &HotelPostgres{db: db}
}

func (h *HotelPostgres) Create(hotel *model.Hotel) (int64, error) {
	if hotel == nil {
		return 0, fmt.Errorf("repo create hotel: %w", my_errors.NilPointerToHotelErr)
	}

	now := time.Now()

	err := h.db.QueryRow(
		`INSERT INTO hotels (name, address, rating, rooms_num, rooms_occupied, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id`,
		hotel.Name,
		hotel.Address,
		hotel.Rating,
		hotel.RoomsNum,
		hotel.RoomsOccupied,
		now,
		now,
	).Scan(&hotel.Id)
	if err != nil {
		return 0, fmt.Errorf("repo create hotel: %w", err)
	}

	return hotel.Id, nil
}

func (h *HotelPostgres) Update(id int64, hotel *model.Hotel) error {
	if hotel == nil {
		return fmt.Errorf("repo update hotel id=%d: %w", id, my_errors.NilPointerToHotelErr)
	}

	result, err := h.db.Exec(
		`UPDATE hotels
		 SET name = $1,
		     address = $2,
		     rating = $3,
		     rooms_num = $4,
		     rooms_occupied = $5,
		     updated_at = $6
		 WHERE id = $7`,
		hotel.Name,
		hotel.Address,
		hotel.Rating,
		hotel.RoomsNum,
		hotel.RoomsOccupied,
		time.Now(),
		id,
	)
	if err != nil {
		return fmt.Errorf("repo update hotel id=%d: %w", id, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("repo update hotel id=%d: rows affected: %w", id, err)
	}

	if rows == 0 {
		return fmt.Errorf("repo update hotel id=%d: %w", id, my_errors.HotelWithIDNotFoundErr)
	}

	return nil
}

func (h *HotelPostgres) Get(id int64) (*model.Hotel, error) {
	var hotel model.Hotel

	err := h.db.QueryRow(
		`SELECT id, created_at, updated_at, name, address, rating, rooms_num, rooms_occupied
		 FROM hotels
		 WHERE id = $1`,
		id,
	).Scan(
		&hotel.Id,
		&hotel.CreatedAt,
		&hotel.UpdatedAt,
		&hotel.Name,
		&hotel.Address,
		&hotel.Rating,
		&hotel.RoomsNum,
		&hotel.RoomsOccupied,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf(
				"repo get hotel id=%d: %w",
				id,
				errors.Join(my_errors.HotelWithIDNotFoundErr, err),
			)
		}
		return nil, fmt.Errorf("repo get hotel id=%d: %w", id, err)
	}

	return &hotel, nil
}

func (h *HotelPostgres) GetAll() ([]*model.Hotel, error) {
	rows, err := h.db.Query(
		`SELECT id, created_at, updated_at, name, address, rating, rooms_num, rooms_occupied
		 FROM hotels`,
	)
	if err != nil {
		return nil, fmt.Errorf("repo get all hotels: %w", err)
	}
	defer rows.Close()

	hotels := make([]*model.Hotel, 0)

	for rows.Next() {
		hotel := new(model.Hotel)

		err = rows.Scan(
			&hotel.Id,
			&hotel.CreatedAt,
			&hotel.UpdatedAt,
			&hotel.Name,
			&hotel.Address,
			&hotel.Rating,
			&hotel.RoomsNum,
			&hotel.RoomsOccupied,
		)
		if err != nil {
			return nil, fmt.Errorf("repo get all hotels: scan row: %w", err)
		}

		hotels = append(hotels, hotel)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("repo get all hotels: rows iteration: %w", err)
	}

	return hotels, nil
}

func (h *HotelPostgres) Delete(id int64) error {
	result, err := h.db.Exec(
		`DELETE FROM hotels WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("repo delete hotel id=%d: %w", id, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("repo delete hotel id=%d: rows affected: %w", id, err)
	}

	if rows == 0 {
		return fmt.Errorf("repo delete hotel id=%d: %w", id, my_errors.HotelWithIDNotFoundErr)
	}

	return nil
}
