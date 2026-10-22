package handler

import (
	"context"
	"encoding/json"
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

func NewHandler(service NotificationService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Handle(ctx context.Context, key, value []byte) error {
	var envelope struct {
		EventType string `json:"event_type"`
	}

	if err := json.Unmarshal(value, &envelope); err != nil {
		return fmt.Errorf("decode event type: %w", err)
	}

	switch envelope.EventType {
	case "booking.created":
		var event events.BookingCreated

		if err := json.Unmarshal(value, &event); err != nil {
			return fmt.Errorf("decode booking.created: %w", err)
		}

		return h.service.BookingCreated(ctx, event)

	case "booking.cancelled":
		var event events.BookingCancelled

		if err := json.Unmarshal(value, &event); err != nil {
			return fmt.Errorf("decode booking.cancelled: %w", err)
		}

		return h.service.BookingCancelled(ctx, event)

	default:
		return fmt.Errorf("unknown event type: %q", envelope.EventType)
	}
}
