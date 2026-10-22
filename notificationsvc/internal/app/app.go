package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/olegetoya/booking/notificationsvc/internal/config"
	"github.com/olegetoya/booking/notificationsvc/internal/handler"
	"github.com/olegetoya/booking/notificationsvc/internal/kafka"
	"github.com/olegetoya/booking/notificationsvc/internal/repository/gmail"
	"github.com/olegetoya/booking/notificationsvc/internal/service"
)

type App struct {
	log      *slog.Logger
	consumer *kafka.Consumer
}

func NewApp(log *slog.Logger, cfg *config.Config) (*App, error) {
	const op = "app.NewApp"

	smtpUser := os.Getenv("SMTP_USER")
	smtpPassword := os.Getenv("SMTP_PASSWORD")

	if smtpUser == "" || smtpPassword == "" {
		return nil, fmt.Errorf("%s: SMTP credentials are empty", op)
	}

	gmailRepo := gmail.NewRepository(
		cfg.SMTP.Host,
		cfg.SMTP.Port,
		smtpUser,
		smtpPassword,
	)

	notificationService := service.NewNotificationService(
		gmailRepo,
		cfg.SMTP.Recipient,
	)

	notificationHandler := handler.NewHandler(notificationService)

	consumer, err := kafka.NewConsumer(
		cfg.Kafka.Brokers,
		cfg.Kafka.Topic,
		cfg.Kafka.GroupID,
		notificationHandler,
		log,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: create consumer: %w", op, err)
	}

	return &App{
		log:      log,
		consumer: consumer,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	a.log.Info("starting notification consumer")

	if err := a.consumer.Run(ctx); err != nil {
		return fmt.Errorf("run consumer: %w", err)
	}

	return nil
}

func (a *App) Stop() {
	a.log.Info("stopping notification consumer")
	a.consumer.Close()
}
