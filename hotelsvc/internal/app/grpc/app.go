package grpcapp

import (
	"fmt"
	grpchandler "github.com/olegetoya/booking/hotelsvc/internal/handler/grpc"
	"google.golang.org/grpc"
	"log/slog"
	"net"
)

type App struct {
	log        *slog.Logger
	gRPCserver *grpc.Server
	port       int
}

func New(log *slog.Logger, rooms grpchandler.Rooms, port int) *App {
	gRPCServer := grpc.NewServer()
	grpchandler.Register(gRPCServer, rooms)
	return &App{log, gRPCServer, port}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
	return
}

func (a *App) Run() error {
	const op = "grpcapp.Run"

	log := a.log.With(
		slog.String("op", op),
		slog.Int("port", a.port),
	)

	log.Info("starting grpc server")

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := a.gRPCserver.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) Stop() {
	const op = "grpcapp.Stop"

	log := a.log.With(slog.String("op", op))
	log.Info("stopping grpc server", slog.Int("port", a.port))

	a.gRPCserver.GracefulStop()

	log.Info("stopped grpc server", slog.Int("port", a.port))

	return
}
