package app

import (
	"fmt"
	"log/slog"
	grpcapp "sso/internal/app/grpc"
	"sso/internal/config"
	"sso/internal/services/auth"
	"sso/internal/storage/postgres"
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
	pgconf := config.LoadPostgresConfig()
	db := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		pgconf.User,
		pgconf.Password,
		pgconf.Host,
		pgconf.Port,
		pgconf.DBName,
		pgconf.SSLMode,
	)
	storage, err := postgres.New(db)
	//storage, err := postgres.New("postgres://admin:123@localhost:5432/mydb?sslmode=disable")
	if err != nil {
		panic(err)
	}

	authService := auth.New(log, storage, storage, storage, totenTTL)

	grpcApp := grpcapp.New(log, authService, grpcPort)

	return &App{GRPCSrv: grpcApp}
}
