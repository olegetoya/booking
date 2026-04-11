package model

import "time"

type RoomType string

const (
	Economy  RoomType = "economic"
	Standard RoomType = "standard"
	Luxury   RoomType = "luxury"
)

type Availability int

const (
	Reserved  Availability = 0
	Pending   Availability = 1
	Available Availability = 2
)

type Room struct {
	Id          int64
	HotelId     int64
	RoomNum     int
	Type        RoomType
	Cost        int
	IsAvailable Availability
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
