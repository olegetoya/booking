package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/olegetoya/booking/notificationsvc/internal/events"
)

type NotificationService interface {
	BookingCreated(ctx context.Context, event events.BookingCreated) error
	BookingCancelled(ctx context.Context, event events.BookingCancelled) error
}

type Handler struct {
	service NotificationService
}

var ErrInvalidEvent = errors.New("invalid kafka event")

func NewHandler(service NotificationService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Handle(
	ctx context.Context,
	key, value []byte,
) error {
	var envelope struct {
		EventType string `json:"event_type"`
	}

	if err := json.Unmarshal(value, &envelope); err != nil {
		return fmt.Errorf(
			"%w: decode event type: %v",
			ErrInvalidEvent, err,
		)
	}

	switch envelope.EventType {
	case "booking.created":
		var event events.BookingCreated

		if err := json.Unmarshal(value, &event); err != nil {
			return fmt.Errorf(
				"%w: decode booking.created: %v",
				ErrInvalidEvent, err,
			)
		}

		if event.EventID == "" || event.BookingID <= 0 {
			return fmt.Errorf(
				"%w: missing booking.created fields",
				ErrInvalidEvent,
			)
		}

		return h.service.BookingCreated(ctx, event)

	case "booking.cancelled":
		var event events.BookingCancelled

		if err := json.Unmarshal(value, &event); err != nil {
			return fmt.Errorf(
				"%w: decode booking.cancelled: %v",
				ErrInvalidEvent, err,
			)
		}

		if event.EventID == "" || event.BookingID <= 0 {
			return fmt.Errorf(
				"%w: missing booking.cancelled fields",
				ErrInvalidEvent,
			)
		}

		return h.service.BookingCancelled(ctx, event)

	default:
		return fmt.Errorf(
			"%w: unknown event type: %q",
			ErrInvalidEvent,
			envelope.EventType,
		)
	}
}
