package domain

type AvailableRoom struct {
	ID           int64
	HotelID      int64
	RoomNum      int64
	Type         string
	CostPerNight int64
	TotalCost    int64
	IsAvailable  bool
}
