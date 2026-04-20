package grpcapp

import (
	"fmt"
	authgRPC "github.com/olegetoya/booking/sso/internal/grpc/auth"
	"google.golang.org/grpc"
	"log/slog"
	"net"
)

type App struct {
	log        *slog.Logger
	gRPCserver *grpc.Server
	port       int
}

func New(log *slog.Logger, auth authgRPC.Auth, port int) *App {
	gRPCServer := grpc.NewServer()
	authgRPC.Register(gRPCServer, auth)
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

	log.Info("grpc server started", slog.String("address", l.Addr().String()))

	if err := a.gRPCserver.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) Stop() {
	const op = "grpcapp.Stop"

	a.log.With(slog.String("op", op))
	a.log.Info("stopping grpc server", slog.Int("port", a.port))

	a.gRPCserver.GracefulStop()

	a.log.Info("stopped grpc server", slog.Int("port", a.port))

	return
}
