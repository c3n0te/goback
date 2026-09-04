package main

import (
	"api"
	"context"
	"log/slog"

	"github.com/jmoiron/sqlx"
)

type GatewayServer struct {
	api.GatewayServer
	DB *sqlx.DB
}

func (srv *GatewayServer) Register(ctx context.Context, req *api.RegRequest) (*api.RegResponse, error) {
	slog.Info("Registering Device")
	regres := &api.RegResponse{
		Registered: false,
	}
	return regres, nil
}
