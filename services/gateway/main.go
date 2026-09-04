package main

import (
	"api"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewGrpcClient(cfg *Config) api.GatewayClient {
	conn, err := grpc.NewClient(
		fmt.Sprintf("%v:%v", cfg.GrpcIp, cfg.GrpcPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		slog.Error("Failed to create new gRPC client")
		os.Exit(1)
	}
	defer conn.Close()

	gbc := api.NewGatewayClient(conn)
	return gbc
}

func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /items", getItems)
	mux.HandleFunc("GET /items/{id}", getItemById)
	mux.HandleFunc("POST /items", createItem)
	return mux
}

func runHttp(mux *http.ServeMux, addr string) {
	slog.Info(fmt.Sprintf("HTTP server running on %v", addr))
	if err := http.ListenAndServe(addr, mux); err != nil {
		slog.Error(fmt.Sprintf("Failed to listen on %v", addr))
		panic(err)
	}
}

func main() {
	cfg := NewConfig()
	mux := NewRouter()
	go runHttp(mux, fmt.Sprintf("%v:%v", cfg.HttpIp, cfg.HttpPort))
	gbc := NewGrpcClient(cfg)
	CallRegister(gbc, cfg)
}
