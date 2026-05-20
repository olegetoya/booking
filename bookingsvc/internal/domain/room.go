package domain

type Room struct {
	ID          int64
	HotelID     int64
	RoomNum     int64
	Type        string
	Cost        int64
	IsAvailable bool
}
