package main

import (
	"github.com/olegetoya/booking/hotelsvc/internal/app"
	"github.com/olegetoya/booking/hotelsvc/internal/config"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()
	logger := setupLogger(cfg.Env)

	slog.Info("starting application", slog.Any("config", cfg))

	application := app.New(logger, cfg.GRPC.Port, cfg.HTTP.Port, cfg.Database.DSN, cfg.HTTP.ReadHeaderTimeout)

	go application.GRPCSrv.MustRun()
	go application.HTTPSrv.MustRun()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	application.GRPCSrv.Stop()
	application.HTTPSrv.Stop()
	application.DB.Close()
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
	return log
}
