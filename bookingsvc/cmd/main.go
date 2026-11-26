package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/olegetoya/booking/bookingsvc/internal/app"
	"github.com/olegetoya/booking/bookingsvc/internal/config"
	"github.com/olegetoya/booking/bookingsvc/internal/telemetry"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	_ = godotenv.Load()
	cfg := config.MustLoad()
	logger := setupLogger(cfg.Env)

	shutdownTracing, err := telemetry.Init(context.Background(), logger, "bookingsvc")
	if err != nil {
		panic(err)
	}
	defer shutdownTracing()

	application, err := app.NewApp(logger, cfg)
	if err != nil {
		logger.Error("failed to initialize app", slog.Any("error", err))
		panic(err)
	}

	go func() {
		logger.Info("starting application bookingsvc")
		if err := application.Run(); err != nil {
			logger.Error("failed to run app:", slog.Any("error", err))
			panic(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := application.Stop(ctx); err != nil {
		logger.Error("failed to stop app:", slog.Any("error", err))
	}
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}
	if log == nil {
		log = slog.Default()
	}
	return slog.New(telemetry.LogHandler{Handler: log.Handler()})
}
