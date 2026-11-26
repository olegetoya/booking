package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/olegetoya/booking/notificationsvc/internal/events"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type NotificationRepository interface {
	Send(
		ctx context.Context,
		to string,
		subject string,
		body string,
		attempt int,
	) error
}

type NotificationService struct {
	repo  NotificationRepository
	email string
}

func NewNotificationService(
	repo NotificationRepository,
	email string,
) *NotificationService {
	return &NotificationService{
		repo:  repo,
		email: email,
	}
}

func (s *NotificationService) BookingCreated(
	ctx context.Context,
	event events.BookingCreated,
) (resultErr error) {
	const op = "service.NotificationService.BookingCreated"
	trace.SpanFromContext(ctx).SetAttributes(attribute.String("event.name", "booking.created"), attribute.Int64("booking.id", event.BookingID))
	ctx, span := otel.Tracer("notificationsvc/service").Start(ctx, "booking.created notify",
		trace.WithAttributes(attribute.Int64("booking.id", event.BookingID)),
	)
	defer func() { finishNotificationSpan(span, resultErr) }()

	subject := fmt.Sprintf(
		"Бронирование №%d подтверждено",
		event.BookingID,
	)

	body := fmt.Sprintf(
		"Ваше бронирование успешно создано!\n\n"+
			"Номер бронирования: %d\n"+
			"Отель: %d\n"+
			"Номер комнаты: %d\n"+
			"Дата заезда: %s\n"+
			"Дата выезда: %s\n"+
			"Общая стоимость: %d\n",
		event.BookingID,
		event.HotelID,
		event.RoomID,
		event.DateFrom.Format("02.01.2006"),
		event.DateTo.Format("02.01.2006"),
		event.TotalCost,
	)

	if err := s.sendWithRetry(ctx, s.email, subject, body); err != nil {
		return fmt.Errorf("%s: send email: %w", op, err)
	}

	return nil
}

func (s *NotificationService) BookingCancelled(
	ctx context.Context,
	event events.BookingCancelled,
) (resultErr error) {
	const op = "service.NotificationService.BookingCancelled"
	trace.SpanFromContext(ctx).SetAttributes(attribute.String("event.name", "booking.cancelled"), attribute.Int64("booking.id", event.BookingID))
	ctx, span := otel.Tracer("notificationsvc/service").Start(ctx, "booking.cancelled notify",
		trace.WithAttributes(attribute.Int64("booking.id", event.BookingID)),
	)
	defer func() { finishNotificationSpan(span, resultErr) }()

	subject := fmt.Sprintf(
		"Бронирование №%d отменено",
		event.BookingID,
	)

	body := fmt.Sprintf(
		"Ваше бронирование №%d успешно отменено.",
		event.BookingID,
	)

	if err := s.repo.Send(ctx, s.email, subject, body, 1); err != nil {
		return fmt.Errorf("%s: send email: %w", op, err)
	}

	return nil
}

func finishNotificationSpan(span trace.Span, err error) {
	defer span.End()
	if err != nil {
		span.RecordError(errors.New("notification delivery failed"))
		span.SetStatus(codes.Error, "notification delivery failed")
	}
}
