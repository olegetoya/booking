package domain

import "time"

type Booking struct {
	ID       int64
	UserID   int64
	HotelID  int64
	RoomID   int64
	DateFrom time.Time
	DateTo   time.Time
	Status   string
}
