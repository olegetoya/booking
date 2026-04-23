package main

import (
	"context"
	"database/sql"
	"errors"
	roomsv1 "github.com/olegetoya/booking/protos/gen/go/hotelsvc/rooms"
	"google.golang.org/grpc"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/olegetoya/booking/hotelsvc/internal/handler"
	"github.com/olegetoya/booking/hotelsvc/internal/middleware"
	"github.com/olegetoya/booking/hotelsvc/internal/repository/hotelrepo"
	"github.com/olegetoya/booking/hotelsvc/internal/repository/roomrepo"
	"github.com/olegetoya/booking/hotelsvc/internal/router"
	"github.com/olegetoya/booking/hotelsvc/internal/service"
)

const (
	readHeaderTimeout = 5 * time.Second
	shutdownTimeout   = 10 * time.Second
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	slog.SetDefault(logger)
	slog.Info("main started")

	connStr := "user=booking_user password=booking_pass dbname=booking_db sslmode=disable"

	db, err := initDB(connStr)
	if err != nil {
		slog.Error("database init failed", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	hotelRepo := hotelrepo.NewHotelPostgres(db)
	hotelServe := service.NewHotelService(hotelRepo)
	hotelHandle := handler.NewHotelHandler(hotelServe)

	roomRepo := roomrepo.NewRoomPostgres(db)
	roomServe := service.NewRoomService(roomRepo)
	roomHandle := handler.NewRoomHandler(roomServe)

	mux := router.NewRouter(hotelHandle, roomHandle)
	loggedMux := middleware.AccessLog(logger, mux)

	gRPCServer := grpc.NewServer()
	roomsv1.RegisterRoomsServer(gRPCServer, roomServe)
	server := &http.Server{
		Addr:              ":8080",
		Handler:           loggedMux,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	go func() {
		slog.Info("server started", "addr", ":8080")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed", "err", err)
			server.Close()
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutdown server ...")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	err = server.Shutdown(ctx)
	if err != nil {
		log.Printf("error while shutting down: %v\n", err)
	}

	log.Println("server closed")
}

func initDB(connStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Проверка подключения
	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
