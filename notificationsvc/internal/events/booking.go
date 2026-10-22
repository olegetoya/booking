package events

import "time"

type BookingCreated struct {
	EventID    string    `json:"event_id"`
	EventType  string    `json:"event_type"`
	BookingID  int64     `json:"booking_id"`
	UserID     int64     `json:"user_id"`
	HotelID    int64     `json:"hotel_id"`
	RoomID     int64     `json:"room_id"`
	DateFrom   time.Time `json:"date_from"`
	DateTo     time.Time `json:"date_to"`
	TotalCost  int64     `json:"total_cost"`
	OccurredAt time.Time `json:"occurred_at"`
}

type BookingCancelled struct {
	EventID    string    `json:"event_id"`
	EventType  string    `json:"event_type"`
	BookingID  int64     `json:"booking_id"`
	OccurredAt time.Time `json:"occurred_at"`
}
