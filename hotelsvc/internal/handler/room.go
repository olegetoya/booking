package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"booking/hotelsvc/internal/dto"
	"booking/hotelsvc/internal/my_errors"
)

type RoomHandler struct {
	service RoomService
}

func NewRoomHandler(service RoomService) *RoomHandler {
	return &RoomHandler{service: service}
}

type RoomService interface {
	Create(hotelID int64, room *dto.Room) error
	Update(hotelID, roomID int64, room *dto.Room) error
	GetAll(hotelID int64) ([]*dto.Room, error)
	Get(hotelID, roomID int64) (*dto.Room, error)
	Delete(hotelID, roomID int64) error
}

func (rh *RoomHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	hotelID, err := parseIDs(r, "hotel_id")
	if err != nil {
		slog.Warn("create room: invalid hotel id", slog.String("hotel_id", r.PathValue("hotel_id")), slog.Any("error", err))
		writeError(w, http.StatusBadRequest, "invalid hotel id")
		return
	}

	var room dto.Room
	if err := decodeJSON(r, &room); err != nil {
		slog.Warn("create room: invalid json", slog.Int64("hotel_id", hotelID), slog.Any("error", err))
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if err := rh.service.Create(hotelID, &room); err != nil {
		handleRoomError(w, "create room", hotelID, 0, true, false, err)
		return
	}

	slog.Info("create room: success", slog.Int64("hotel_id", hotelID))
	w.WriteHeader(http.StatusCreated)
}

func (rh *RoomHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	hotelID, err := parseIDs(r, "hotel_id")
	if err != nil {
		slog.Warn("get all rooms: invalid hotel id", slog.String("hotel_id", r.PathValue("hotel_id")), slog.Any("error", err))
		writeError(w, http.StatusBadRequest, "invalid hotel id")
		return
	}

	rooms, err := rh.service.GetAll(hotelID)
	if err != nil {
		handleRoomError(w, "get all rooms", hotelID, 0, true, false, err)
		return
	}

	writeJSON(w, http.StatusOK, rooms)
}

func (rh *RoomHandler) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	hotelID, err := parseIDs(r, "hotel_id")
	if err != nil {
		slog.Warn("get room: invalid hotel id", slog.String("hotel_id", r.PathValue("hotel_id")), slog.Any("error", err))
		writeError(w, http.StatusBadRequest, "invalid hotel id")
		return
	}

	roomID, err := parseIDs(r, "room_id")
	if err != nil {
		slog.Warn("get room: invalid room id", slog.String("room_id", r.PathValue("room_id")), slog.Any("error", err))
		writeError(w, http.StatusBadRequest, "invalid room id")
		return
	}

	room, err := rh.service.Get(hotelID, roomID)
	if err != nil {
		handleRoomError(w, "get room", hotelID, roomID, true, true, err)
		return
	}

	writeJSON(w, http.StatusOK, room)
}

func (rh *RoomHandler) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	hotelID, err := parseIDs(r, "hotel_id")
	if err != nil {
		slog.Warn("update room: invalid hotel id", slog.String("hotel_id", r.PathValue("hotel_id")), slog.Any("error", err))
		writeError(w, http.StatusBadRequest, "invalid hotel id")
		return
	}

	roomID, err := parseIDs(r, "room_id")
	if err != nil {
		slog.Warn("update room: invalid room id", slog.String("room_id", r.PathValue("room_id")), slog.Any("error", err))
		writeError(w, http.StatusBadRequest, "invalid room id")
		return
	}

	var room dto.Room
	if err := decodeJSON(r, &room); err != nil {
		slog.Warn(
			"update room: invalid json",
			slog.Int64("hotel_id", hotelID),
			slog.Int64("room_id", roomID),
			slog.Any("error", err),
		)
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if err := rh.service.Update(hotelID, roomID, &room); err != nil {
		handleRoomError(w, "update room", hotelID, roomID, true, true, err)
		return
	}

	slog.Info("update room: success", slog.Int64("hotel_id", hotelID), slog.Int64("room_id", roomID))
	w.WriteHeader(http.StatusNoContent)
}

func (rh *RoomHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	hotelID, err := parseIDs(r, "hotel_id")
	if err != nil {
		slog.Warn("delete room: invalid hotel id", slog.String("hotel_id", r.PathValue("hotel_id")), slog.Any("error", err))
		writeError(w, http.StatusBadRequest, "invalid hotel id")
		return
	}

	roomID, err := parseIDs(r, "room_id")
	if err != nil {
		slog.Warn("delete room: invalid room id", slog.String("room_id", r.PathValue("room_id")), slog.Any("error", err))
		writeError(w, http.StatusBadRequest, "invalid room id")
		return
	}

	if err := rh.service.Delete(hotelID, roomID); err != nil {
		handleRoomError(w, "delete room", hotelID, roomID, true, true, err)
		return
	}

	slog.Info("delete room: success", slog.Int64("hotel_id", hotelID), slog.Int64("room_id", roomID))
	w.WriteHeader(http.StatusNoContent)
}

func handleRoomError(
	w http.ResponseWriter,
	operation string,
	hotelID, roomID int64,
	withHotelID, withRoomID bool,
	err error,
) {
	switch {
	case errors.Is(err, my_errors.RoomValidationErr):
		attrs := make([]any, 0, 3)
		if withHotelID {
			attrs = append(attrs, slog.Int64("hotel_id", hotelID))
		}
		if withRoomID {
			attrs = append(attrs, slog.Int64("room_id", roomID))
		}
		attrs = append(attrs, slog.Any("error", err))
		slog.Warn(operation+": validation failed", attrs...)
		writeError(w, http.StatusBadRequest, "invalid room data")

	case errors.Is(err, my_errors.HotelWithIDNotFoundErr):
		attrs := make([]any, 0, 1)
		if withHotelID {
			attrs = append(attrs, slog.Int64("hotel_id", hotelID))
		}
		slog.Info(operation+": hotel not found", attrs...)
		writeError(w, http.StatusNotFound, "hotel not found")

	case errors.Is(err, my_errors.RoomWithIDNotFoundErr):
		attrs := make([]any, 0, 2)
		if withHotelID {
			attrs = append(attrs, slog.Int64("hotel_id", hotelID))
		}
		if withRoomID {
			attrs = append(attrs, slog.Int64("room_id", roomID))
		}
		slog.Info(operation+": room not found", attrs...)
		writeError(w, http.StatusNotFound, "room not found")

	default:
		attrs := make([]any, 0, 3)
		if withHotelID {
			attrs = append(attrs, slog.Int64("hotel_id", hotelID))
		}
		if withRoomID {
			attrs = append(attrs, slog.Int64("room_id", roomID))
		}
		attrs = append(attrs, slog.Any("error", err))
		slog.Error(operation+": internal error", attrs...)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

func parseIDs(r *http.Request, value string) (int64, error) {
	idStr := r.PathValue(value)

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, err
	}
	if id <= 0 {
		return 0, my_errors.InvalidIDErr
	}

	return id, nil
}
