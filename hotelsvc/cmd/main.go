package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/olegetoya/booking/hotelsvc/internal/app"
	"github.com/olegetoya/booking/hotelsvc/internal/config"
	"github.com/olegetoya/booking/hotelsvc/internal/telemetry"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	_ = godotenv.Load()
	dsn := os.Getenv("DB_DSN")

	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)

	shutdownTracing, err := telemetry.Init(context.Background(), log, "hotelsvc")
	if err != nil {
		panic(err)
	}
	defer shutdownTracing()

	log.Info("starting application hotelsvc", slog.Any("config", cfg))

	application := app.New(log, cfg.GRPC.Port, cfg.HTTP.Port, dsn, cfg.HTTP.ReadHeaderTimeout)

	go application.GRPCSrv.MustRun()
	go application.HTTPSrv.MustRun()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	application.GRPCSrv.Stop()
	err = application.HTTPSrv.Stop()

	if err != nil {
		log.Error("failed to stop http server", slog.String("error", err.Error()))
	}
	err = application.DB.Close()
	if err != nil {
		log.Error("failed to close database", slog.String("error", err.Error()))
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
