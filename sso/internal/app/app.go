package app

import (
	"log/slog"
	grpcapp "sso/internal/app/grpc"
	"time"
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(
	log *slog.Logger,
	grpcPort int,
	storagepath string,
	totenTTL time.Duration,
) *App {
	//TODO: storage
	//TODO: init auth service(auth)

	grpcApp := grpcapp.New(log, grpcPort)

	return &App{GRPCSrv: grpcApp}
}
