package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	hotelsgrpc "github.com/olegetoya/booking/bookingsvc/internal/clients/hotelsvc/grpc"
	"github.com/olegetoya/booking/bookingsvc/internal/config"
	bookinggen "github.com/olegetoya/booking/bookingsvc/internal/gen/bookingv1"
	bookinghandler "github.com/olegetoya/booking/bookingsvc/internal/handler"
	"github.com/olegetoya/booking/bookingsvc/internal/service"
	"github.com/olegetoya/booking/bookingsvc/internal/storage/postgres"

	_ "github.com/lib/pq"
)

type App struct {
	log        *slog.Logger
	httpServer *http.Server
	db         *sql.DB
}

func NewApp(log *slog.Logger, cfg *config.Config) (*App, error) {
	const op = "app.NewApp"

	postgresDSN := os.Getenv("POSTGRES_DSN")
	if postgresDSN == "" {
		return nil, fmt.Errorf("%s: POSTGRES_DSN is empty", op)
	}

	db, err := sql.Open("postgres", postgresDSN)
	if err != nil {
		return nil, fmt.Errorf("%s: open postgres: %w", op, err)
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("%s: ping postgres: %w", op, err)
	}

	bookingRepo := postgres.NewBookingRepository(db)

	hotelsAddr := net.JoinHostPort(
		cfg.Clients.HotelSvc.GRPC.Host,
		cfg.Clients.HotelSvc.GRPC.Port,
	)

	roomsClient, err := hotelsgrpc.NewRoomsClient(
		hotelsAddr,
		cfg.Clients.HotelSvc.GRPC.Timeout,
	)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("%s: create hotels grpc client: %w", op, err)
	}

	bookingService := service.NewBookingService(
		log,
		bookingRepo,
		roomsClient,
	)

	bookingHandler := bookinghandler.NewHandler(bookingService)

	bookingServer, err := bookinggen.NewServer(bookingHandler)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("%s: create openapi server: %w", op, err)
	}

	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(10 * time.Second))

	router.Mount("/api/v1", bookingServer)

	httpServer := &http.Server{
		Addr:              net.JoinHostPort(cfg.HTTP.Host, cfg.HTTP.Port),
		Handler:           router,
		ReadHeaderTimeout: cfg.HTTP.ReadHeaderTimeout,
		ReadTimeout:       cfg.HTTP.ReadTimeout,
		WriteTimeout:      cfg.HTTP.WriteTimeout,
		IdleTimeout:       cfg.HTTP.IdleTimeout,
	}

	return &App{
		log:        log,
		httpServer: httpServer,
		db:         db,
	}, nil
}

func (a *App) Run() error {
	const op = "app.App.Run"

	a.log.Info(
		"starting http server",
		slog.String("addr", a.httpServer.Addr),
	)

	err := a.httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("%s: listen and serve: %w", op, err)
	}

	return nil
}

func (a *App) Stop(ctx context.Context) error {
	const op = "app.App.Stop"

	log := a.log.With(slog.String(op, "App.Stop"))

	log.Info("stopping http server", slog.String("port", a.httpServer.Addr))

	var resultErr error

	if err := a.httpServer.Shutdown(ctx); err != nil {
		log.Error("failed to stop http server", slog.String("port", a.httpServer.Addr), slog.String("err", err.Error()))
		resultErr = fmt.Errorf("%s: shutdown http server: %w", op, err)
	}

	log.Info("stopped http server", slog.String("port", a.httpServer.Addr))

	if err := a.db.Close(); err != nil {
		if resultErr != nil {
			resultErr = fmt.Errorf("%v; close db: %w", resultErr, err)
		} else {
			resultErr = fmt.Errorf("%s: close db: %w", op, err)
		}
	}

	return resultErr
}
