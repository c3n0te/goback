package main

import (
	"api"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewGrpcClientConnWithRetry(cfg *Config) *grpc.ClientConn {
	var conn *grpc.ClientConn
	var err error

	for range 5 {
		conn, err = grpc.NewClient(
			fmt.Sprintf("%v:%v", cfg.GrpcIp, cfg.GrpcPort),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)

		if err != nil {
			slog.Error("Failed to connect to gRPC server: %v", "error", err)
			slog.Info("gRPC server not ready, sleeping for 3 seconds")
			time.Sleep(3 * time.Second)
			continue
		}

		slog.Info("gRPC Client ready")
		break
	}

	if conn == nil {
		slog.Error("Failed to connect to gRPC server")
		os.Exit(1)
	}

	return conn
}

func main() {
	cfg := NewConfig()

	conn := NewGrpcClientConnWithRetry(cfg)
	defer conn.Close()
	gbc := api.NewGoBackClient(conn)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /account/{userid}", func(w http.ResponseWriter, r *http.Request) {
		userId := r.PathValue("userid")
		accres, err := CallGetAccount(gbc, userId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(accres)
	})

	mux.HandleFunc("GET /quote/{ticker}", func(w http.ResponseWriter, r *http.Request) {
		ticker := r.PathValue("ticker")
		qres, err := CallGetQuote(gbc, ticker)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(qres)
	})

	mux.HandleFunc("GET /transactions/{userid}", func(w http.ResponseWriter, r *http.Request) {
		userId := r.PathValue("userid")
		txres, err := CallGetTransactions(gbc, userId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(txres)
	})

	addr := fmt.Sprintf("%v:%v", cfg.HttpIp, cfg.HttpPort)
	slog.Info(fmt.Sprintf("HTTP server running on %v", addr))
	if err := http.ListenAndServe(addr, mux); err != nil {
		slog.Error(fmt.Sprintf("Failed to listen on %v", addr))
		panic(err)
	}
}
