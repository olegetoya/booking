package app

import (
	"database/sql"
	grpcapp "github.com/olegetoya/booking/hotelsvc/internal/app/grpc"
	httpapp "github.com/olegetoya/booking/hotelsvc/internal/app/http"
	"github.com/olegetoya/booking/hotelsvc/internal/handler"
	"github.com/olegetoya/booking/hotelsvc/internal/middleware"
	"github.com/olegetoya/booking/hotelsvc/internal/repository/hotelrepo"
	"github.com/olegetoya/booking/hotelsvc/internal/repository/roomrepo"
	"github.com/olegetoya/booking/hotelsvc/internal/router"
	"github.com/olegetoya/booking/hotelsvc/internal/service"
	grpcSrv "github.com/olegetoya/booking/hotelsvc/internal/service/grpc"
	"log/slog"
	"time"
)

type App struct {
	GRPCSrv *grpcapp.App
	HTTPSrv *httpapp.App
	DB      *sql.DB
}

func New(
	log *slog.Logger,
	grpcPort int,
	httpPort int,
	dsn string,
	readHeaderTimeout time.Duration,
) *App {
	db, err := initDB(dsn)
	if err != nil {
		panic(err)
	}

	hotelRepo := hotelrepo.NewHotelPostgres(db)
	hotelServe := service.NewHotelService(hotelRepo)
	hotelHandle := handler.NewHotelHandler(hotelServe)

	roomRepo := roomrepo.NewRoomPostgres(db)
	roomServe := service.NewRoomService(roomRepo)
	roomHandle := handler.NewRoomHandler(roomServe)

	mux := router.NewRouter(hotelHandle, roomHandle)
	loggedMux := middleware.AccessLog(log, mux)

	roomServeGRPC := grpcSrv.NewRoomsServer(roomRepo)

	httpApp := httpapp.New(log, loggedMux, httpPort, readHeaderTimeout)
	grpcApp := grpcapp.New(log, roomServeGRPC, grpcPort)

	return &App{
		GRPCSrv: grpcApp,
		HTTPSrv: httpApp,
		DB:      db,
	}
}

func initDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
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
