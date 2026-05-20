package domain

import "errors"

var (
	ErrBookingNotFound    = errors.New("booking not found")
	ErrRoomNotFound       = errors.New("room not found")
	ErrRoomAlreadyBooked  = errors.New("room already booked")
	ErrInvalidDateRange   = errors.New("invalid date range")
	ErrRoomFromOtherHotel = errors.New("room does not belong to hotel")
)
