package dto

type gRPCRoom struct {
	Id          int64  `json:"id"`
	HotelId     int64  `json:"hotelId"`
	RoomNum     int    `json:"room_num"`
	Type        string `json:"type"`
	Cost        int    `json:"cost"`
	IsAvailable int    `json:"is_available"`
}
