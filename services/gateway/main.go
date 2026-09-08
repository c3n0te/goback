package main

import (
	"api"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	rmq "github.com/rabbitmq/rabbitmq-amqp-go-client/pkg/rabbitmqamqp"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewQueueWithRetry(cfg *Config) (*rmq.AmqpConnection, *rmq.Environment, *rmq.Publisher) {
	ctx := context.Background()
	env := rmq.NewEnvironment(cfg.QueueUrl, nil)
	var conn *rmq.AmqpConnection
	var publisher *rmq.Publisher
	var err error

	for range 5 {
		conn, err = env.NewConnection(ctx)
		if err != nil {
			slog.Error("Failed to connect to RabbitMQ: ", "error", err)
			slog.Info("Queue Not ready, sleeping for 3 seconds")
			time.Sleep(3 * time.Second)
			continue
		}

		break
	}

	_, err = conn.Management().DeclareQueue(ctx, &rmq.QuorumQueueSpecification{
		Name: cfg.QueueName,
	})

	if err != nil {
		slog.Error("Failed to declare a queue: ", "error", err)

	}

	publisher, err = conn.NewPublisher(ctx, &rmq.QueueAddress{Queue: cfg.QueueName}, nil)
	if err != nil {
		slog.Error("Failed to create queue publisher: ", "error", err)
	}

	if conn == nil || env == nil || publisher == nil {
		slog.Error("Failed to connect to Queue")
		os.Exit(1)
	}

	slog.Info("Queue ready")
	return conn, env, publisher
}

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

		break
	}

	if conn == nil {
		slog.Error("Failed to connect to gRPC server")
		os.Exit(1)
	}

	slog.Info("gRPC Client ready")
	return conn
}

func main() {
	cfg := NewConfig()

	_, env, qpublisher := NewQueueWithRetry(cfg)
	defer func() {
		_ = env.CloseConnections(context.Background())
		_ = qpublisher.Close(context.Background())
	}()

	conn := NewGrpcClientConnWithRetry(cfg)
	defer conn.Close()
	gbc := api.NewGoBackClient(conn)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /account", func(w http.ResponseWriter, r *http.Request) {
		var newAccount api.CreateAccountRequest
		if err := json.NewDecoder(r.Body).Decode(&newAccount); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		createAccountRes, err := CallCreateAccount(gbc, &newAccount)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(createAccountRes)
	})

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

	mux.HandleFunc("POST /buy", func(w http.ResponseWriter, r *http.Request) {
		var newBuy api.TxRequest
		if err := json.NewDecoder(r.Body).Decode(&newBuy); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		buyRes, err := CallQueuePublishBuy(qpublisher, &newBuy)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(buyRes)
	})

	mux.HandleFunc("POST /sell", func(w http.ResponseWriter, r *http.Request) {
		var newSell api.TxRequest
		if err := json.NewDecoder(r.Body).Decode(&newSell); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		sellRes, err := CallQueuePublishSell(qpublisher, &newSell)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(sellRes)
	})

	mux.HandleFunc("GET /transactions/{userid}", func(w http.ResponseWriter, r *http.Request) {
		userId := r.PathValue("userid")
		txres, err := CallGetTransactionLogs(gbc, userId)
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
