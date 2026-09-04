package main

import (
	"api"
	"fmt"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"google.golang.org/grpc"
)

type GatewayServer struct {
	api.GatewayServer
	DB *sqlx.DB
}

func main() {
	cfg := NewConfig()

	var (
		db  *sqlx.DB
		err error
	)

	for range 5 {
		db, err = sqlx.Connect(cfg.DbType, cfg.DbUrl)
		if err != nil {
			slog.Error("Failed to connect to DB", "error", err)
			os.Exit(1)
		}

		slog.Info("DB not ready, sleeping for 3 seconds")
		time.Sleep(3 * time.Second)
	}

	srv := grpc.NewServer()
	gateway := GatewayServer{
		DB: db,
	}

	api.RegisterGatewayServer(srv, &gateway)
	listener, err := net.Listen(
		"tcp",
		fmt.Sprintf("%v:%v", cfg.GrpcIp, cfg.GrpcPort),
	)

	if err != nil {
		slog.Error("Failed to listen to socket")
		os.Exit(1)
	}

	slog.Info(fmt.Sprintf("gRPC Server running on %v:%v", cfg.GrpcIp, cfg.GrpcPort))
	err = srv.Serve(listener)
	if err != nil {
		slog.Error("Failed to bind server to listener")
		os.Exit(1)
	}
}
