package router

import (
	"net/http"

	"github.com/olegetoya/booking/hotelsvc/internal/handler"
)

func NewRouter(hotelHandler *handler.HotelHandler, roomHandler *handler.RoomHandler) *http.ServeMux {
	mux := http.NewServeMux()
	// Hotels
	mux.HandleFunc("POST /hotels", hotelHandler.Create)
	mux.HandleFunc("GET /hotels", hotelHandler.GetAll)
	mux.HandleFunc("GET /hotels/{id}", hotelHandler.Get)
	mux.HandleFunc("PUT /hotels/{id}", hotelHandler.Update)
	mux.HandleFunc("DELETE /hotels/{id}", hotelHandler.Delete)
	// Rooms
	mux.HandleFunc("POST /hotels/{hotel_id}/rooms", roomHandler.Create)
	mux.HandleFunc("GET /hotels/{hotel_id}/rooms/{room_id}", roomHandler.Get)
	mux.HandleFunc("GET /hotels/{hotel_id}/rooms", roomHandler.GetAll)
	mux.HandleFunc("PUT /hotels/{hotel_id}/rooms/{room_id}", roomHandler.Update)
	mux.HandleFunc("DELETE /hotels/{hotel_id}/rooms/{room_id}", roomHandler.Delete)

	return mux
}
