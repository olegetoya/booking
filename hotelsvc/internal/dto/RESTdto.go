package dto

type Hotel struct {
	Rating        int    `json:"rating"`
	Name          string `json:"name"`
	Address       string `json:"address"`
	RoomsNum      int    `json:"rooms_num"`
	RoomsOccupied int    `json:"rooms_occupied"`
}

type Room struct {
	RoomNum     int    `json:"room_num"`
	Type        string `json:"type"`
	Cost        int    `json:"cost"`
	IsAvailable int    `json:"is_available"`
}
