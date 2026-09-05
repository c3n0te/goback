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

func NewDbWithRetry(cfg *Config) *sqlx.DB {
	var db *sqlx.DB
	var err error

	slog.Info(fmt.Sprintf("DB URL: %v", cfg.DbUrl))
	for range 5 {
		db, err = sqlx.Connect(cfg.DbType, cfg.DbUrl)
		if err != nil {
			slog.Error("Failed to connect to DB", "error", err)
			slog.Info("DB Not ready, sleeping for 3 seconds")
			time.Sleep(3 * time.Second)
			continue
		}
		break
	}
	return db
}

func main() {
	cfg := NewConfig()

	db := NewDbWithRetry(cfg)
	defer db.Close()
	Migrate(db)

	srv := grpc.NewServer()
	goback := NewGatewayServer(db)
	api.RegisterGoBackServer(srv, &goback)
	listener, err := net.Listen(
		"tcp",
		fmt.Sprintf("%v:%v", cfg.GrpcIp, cfg.GrpcPort),
	)

	if err != nil {
		slog.Error("Failed to listen to socket: ", "error", err)
		os.Exit(1)
	}

	slog.Info(fmt.Sprintf("gRPC Server running on %v:%v", cfg.GrpcIp, cfg.GrpcPort))
	err = srv.Serve(listener)
	if err != nil {
		slog.Error("Failed to bind server to listener: ", "error", err)
		os.Exit(1)
	}
}
