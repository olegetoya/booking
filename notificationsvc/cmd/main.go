package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/olegetoya/booking/notificationsvc/internal/app"
	"github.com/olegetoya/booking/notificationsvc/internal/config"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	if err := run(log); err != nil {
		log.Error("application stopped", slog.Any("error", err))
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	cfg := config.MustLoad()

	application, err := app.NewApp(log, cfg)
	if err != nil {
		return fmt.Errorf("initialize application: %w", err)
	}
	defer application.Stop()

	mux := http.NewServeMux()
	mux.Handle("GET /metrics", promhttp.Handler())

	metricsServer := &http.Server{
		Addr:              ":8082",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	listener, err := net.Listen("tcp", metricsServer.Addr)
	if err != nil {
		return fmt.Errorf("listen metrics: %w", err)
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	serverErrors := make(chan error, 1)

	go func() {
		if err := metricsServer.Serve(listener); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
			stop()
		}
	}()

	defer func() {
		shutdownCtx, cancel := context.WithTimeout(
			context.Background(), 5*time.Second,
		)
		defer cancel()

		if err := metricsServer.Shutdown(shutdownCtx); err != nil {
			log.Error("metrics shutdown failed", slog.Any("error", err))
		}
	}()

	if err := application.Run(ctx); err != nil {
		return fmt.Errorf("run consumer: %w", err)
	}

	select {
	case err := <-serverErrors:
		return fmt.Errorf("metrics server: %w", err)
	default:
	}

	return nil
}
