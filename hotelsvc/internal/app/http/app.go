package httpapp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type App struct {
	log        *slog.Logger
	HTTPServer *http.Server
	port       int
}

func New(log *slog.Logger, mux http.Handler, port int, readHeaderTimeout time.Duration) *App {
	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           mux,
		ReadHeaderTimeout: readHeaderTimeout,
	}
	return &App{log: log, HTTPServer: server, port: port}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "httpapp.Run"

	log := a.log.With(
		slog.String("op", op),
		slog.Int("port", a.port),
	)

	log.Info("starting http server")
	err := a.HTTPServer.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (a *App) Stop() error {
	const op = "httpapp.Stop"

	log := a.log.With(
		slog.String("op", op),
		slog.Int("port", a.port),
	)
	log.Info("stopping http server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := a.HTTPServer.Shutdown(shutdownCtx); err != nil {
		log.Error("graceful shutdown failed", "error", err)

		if closeErr := a.HTTPServer.Close(); closeErr != nil {
			log.Error("force close failed", "error", closeErr)
		}
		return fmt.Errorf("%s: shutdown: %w", op, err)
	}

	log.Info("stopped http server")

	return nil
}
