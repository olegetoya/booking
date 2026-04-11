package roomrepo

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"booking/hotelsvc/internal/model"
	"booking/hotelsvc/internal/my_errors"
	_ "github.com/lib/pq"
)

type RoomPostgres struct {
	db *sql.DB
}

func NewRoomPostgres(db *sql.DB) *RoomPostgres {
	return &RoomPostgres{db: db}
}

func (r *RoomPostgres) Create(hotelId int64, room *model.Room) (int64, error) {
	err := r.db.QueryRow(
		`INSERT INTO rooms (hotel_id, room_num, type, cost, available, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
    RETURNING id`,
		hotelId,
		room.RoomNum,
		room.Type,
		room.Cost,
		room.IsAvailable,
		time.Now(),
		time.Now(),
	).Scan(&room.Id)
	if err != nil {
		return 0, fmt.Errorf("repo create room: %w", err)
	}
	return room.Id, nil
}

func (r *RoomPostgres) Update(hotelId, id int64, room *model.Room) error {
	result, err := r.db.Exec(
		`UPDATE rooms 
SET
		hotel_id = $1,
		room_num = $2,
		type = $3,
		available = $4,
		created_at = $5,
		updated_at = $6
WHERE id = $7`,
		hotelId,
		room.RoomNum,
		room.Type,
		room.IsAvailable,
		room.CreatedAt,
		time.Now(),
		id)
	if err != nil {
		return fmt.Errorf("repo update room id=%d: %w", id, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("repo update room id=%d: rows affected: %w", id, err)
	}

	if rows == 0 {
		return fmt.Errorf("repo update room id=%d: %w", id, my_errors.RoomWithIDNotFoundErr)
	}

	return nil
}

func (r *RoomPostgres) Get(hotelId, id int64) (*model.Room, error) {
	var room model.Room
	err := r.db.QueryRow(
		`SELECT hotel_id, room_num, type, cost, available, created_at, updated_at FROM rooms WHERE id = $1 AND hotel_id = $2`,
		id, hotelId).Scan(&room.HotelId, &room.RoomNum, &room.Type, &room.Cost, &room.IsAvailable, &room.CreatedAt, &room.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf(
				"repo get room id=%d: %w",
				id,
				errors.Join(my_errors.RoomWithIDNotFoundErr, err),
			)
		}
		return nil, fmt.Errorf("repo get room id=%d: %w", id, err)
	}
	return &room, nil
}

func (r *RoomPostgres) GetAll(hotelId int64) ([]*model.Room, error) {
	var rooms []*model.Room

	rows, err := r.db.Query(
		`SELECT id, hotel_id, room_num, type, cost, available, created_at, updated_at FROM rooms WHERE hotel_id = $1`, hotelId,
	)
	if err != nil {
		return nil, fmt.Errorf("repo get rooms: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var room model.Room
		err = rows.Scan(
			&room.Id,
			&room.HotelId,
			&room.RoomNum,
			&room.Type,
			&room.Cost,
			&room.IsAvailable,
			&room.CreatedAt,
			&room.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("repo get rooms: scan row: %w", err)
		}
		rooms = append(rooms, &room)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("repo get all rooms: rows iteration: %w", err)
	}
	return rooms, nil
}

func (r *RoomPostgres) Delete(hotelId, id int64) error {
	result, err := r.db.Exec(
		`DELETE FROM rooms WHERE id = $1 AND hotel_id = $2`, id, hotelId)
	if err != nil {
		return fmt.Errorf("repo delete room id=%d: %w", id, err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("repo delete room id=%d: rows affected: %w", id, err)
	}
	if rows == 0 {
		return fmt.Errorf("repo delete room id=%d: %w", id, my_errors.RoomWithIDNotFoundErr)
	}
	return nil
}
