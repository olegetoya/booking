package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/olegetoya/booking/hotelsvc/internal/dto"
	"github.com/olegetoya/booking/hotelsvc/internal/my_errors"
)

type HotelHandler struct {
	service HotelService
}

func NewHotelHandler(service HotelService) *HotelHandler {
	return &HotelHandler{service: service}
}

type HotelService interface {
	Create(hotel *dto.Hotel) error
	Update(id int64, hotel *dto.Hotel) error
	GetAll() ([]*dto.Hotel, error)
	Get(id int64) (*dto.Hotel, error)
	Delete(id int64) error
}

func (h *HotelHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var hotel dto.Hotel
	if err := decodeJSON(r, &hotel); err != nil {
		slog.Warn("create hotel: invalid json", slog.Any("error", err))
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if err := h.service.Create(&hotel); err != nil {
		handleHotelError(w, "create hotel", 0, false, err)
		return
	}

	slog.Info("create hotel: success")
	w.WriteHeader(http.StatusCreated)
}

func (h *HotelHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	hotels, err := h.service.GetAll()
	if err != nil {
		slog.Error("get all hotels: internal error", slog.Any("error", err))
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, hotels)
}

func (h *HotelHandler) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	id, err := parseID(r)
	if err != nil {
		slog.Warn("get hotel: invalid id", slog.String("id", r.PathValue("id")), slog.Any("error", err))
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	hotel, err := h.service.Get(id)
	if err != nil {
		handleHotelError(w, "get hotel", id, true, err)
		return
	}

	writeJSON(w, http.StatusOK, hotel)
}

func (h *HotelHandler) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	id, err := parseID(r)
	if err != nil {
		slog.Warn("update hotel: invalid id", slog.String("id", r.PathValue("id")), slog.Any("error", err))
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var hotel dto.Hotel
	if err := decodeJSON(r, &hotel); err != nil {
		slog.Warn("update hotel: invalid json", slog.Int64("id", id), slog.Any("error", err))
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if err := h.service.Update(id, &hotel); err != nil {
		handleHotelError(w, "update hotel", id, true, err)
		return
	}

	slog.Info("update hotel: success", slog.Int64("id", id))
	w.WriteHeader(http.StatusNoContent)
}

func (h *HotelHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	id, err := parseID(r)
	if err != nil {
		slog.Warn("delete hotel: invalid id", slog.String("id", r.PathValue("id")), slog.Any("error", err))
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.service.Delete(id); err != nil {
		handleHotelError(w, "delete hotel", id, true, err)
		return
	}

	slog.Info("delete hotel: success", slog.Int64("id", id))
	w.WriteHeader(http.StatusNoContent)
}

func handleHotelError(w http.ResponseWriter, operation string, id int64, withID bool, err error) {
	switch {
	case errors.Is(err, my_errors.HotelValidationErr):
		if withID {
			slog.Warn(operation+": validation failed", slog.Int64("id", id), slog.Any("error", err))
		} else {
			slog.Warn(operation+": validation failed", slog.Any("error", err))
		}
		writeError(w, http.StatusBadRequest, "invalid hotel data")

	case errors.Is(err, my_errors.HotelWithIDNotFoundErr):
		if withID {
			slog.Info(operation+": hotel not found", slog.Int64("id", id))
		} else {
			slog.Info(operation + ": hotel not found")
		}
		writeError(w, http.StatusNotFound, "hotel not found")

	default:
		if withID {
			slog.Error(operation+": internal error", slog.Int64("id", id), slog.Any("error", err))
		} else {
			slog.Error(operation+": internal error", slog.Any("error", err))
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

func decodeJSON(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return err
	}

	// Не даём прислать второй JSON-объект в body.
	if decoder.More() {
		return errors.New("request body must contain only one json object")
	}

	return nil
}

func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(data); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, _ = w.Write(buf.Bytes())
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, map[string]string{
		"error": message,
	})
}

func parseID(r *http.Request) (int64, error) {
	idStr := r.PathValue("id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, err
	}
	if id <= 0 {
		return 0, my_errors.InvalidIDErr
	}

	return id, nil
}
