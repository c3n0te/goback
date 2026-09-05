package main

import (
	"api"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	cfg := NewConfig()
	slog.Info("Sleeping to give gRPC server time to boot up")
	time.Sleep(5 * time.Second)
	conn, err := grpc.NewClient(
		fmt.Sprintf("%v:%v", cfg.GrpcIp, cfg.GrpcPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		slog.Error("Failed to create new gRPC client", "error", err)
		os.Exit(1)
	}
	defer conn.Close()

	gbc := api.NewGoBackClient(conn)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /account", func(w http.ResponseWriter, r *http.Request) {
		CallGetAccount(gbc)
	})

	mux.HandleFunc("GET /transactions", func(w http.ResponseWriter, r *http.Request) {
		CallGetTransactions(gbc)
	})

	addr := fmt.Sprintf("%v:%v", cfg.HttpIp, cfg.HttpPort)
	slog.Info(fmt.Sprintf("HTTP server running on %v", addr))
	if err := http.ListenAndServe(addr, mux); err != nil {
		slog.Error(fmt.Sprintf("Failed to listen on %v", addr))
		panic(err)
	}
}
