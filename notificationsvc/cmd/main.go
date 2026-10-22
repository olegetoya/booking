package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/olegetoya/booking/notificationsvc/internal/app"
	"github.com/olegetoya/booking/notificationsvc/internal/config"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cfg := config.MustLoad()

	application, err := app.NewApp(log, cfg)
	if err != nil {
		log.Error("failed to initialize application",
			slog.Any("error", err))
		os.Exit(1)
	}
	defer application.Stop()

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	if err := application.Run(ctx); err != nil {
		log.Error("application stopped with error",
			slog.Any("error", err))
		os.Exit(1)
	}
}
