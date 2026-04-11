package model

import "time"

type HotelRating int

const (
	NoRating  HotelRating = 0
	ThreeStar HotelRating = 3
	FourStar  HotelRating = 4
	FiveStar  HotelRating = 5
)

type Hotel struct {
	Id            int64
	Rating        HotelRating
	Name          string
	Address       string
	RoomsNum      int
	RoomsOccupied int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
