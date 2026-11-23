package service

import (
	"context"
	"fmt"

	"github.com/olegetoya/booking/notificationsvc/internal/events"
)

type NotificationRepository interface {
	Send(
		ctx context.Context,
		to string,
		subject string,
		body string,
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
) error {
	const op = "service.NotificationService.BookingCreated"

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
) error {
	const op = "service.NotificationService.BookingCancelled"

	subject := fmt.Sprintf(
		"Бронирование №%d отменено",
		event.BookingID,
	)

	body := fmt.Sprintf(
		"Ваше бронирование №%d успешно отменено.",
		event.BookingID,
	)

	if err := s.repo.Send(ctx, s.email, subject, body); err != nil {
		return fmt.Errorf("%s: send email: %w", op, err)
	}

	return nil
}
