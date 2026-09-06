package main

import (
	"api"
	"context"
	"log/slog"
	"time"
	"uuid"

	"github.com/jmoiron/sqlx"
	rmq "github.com/rabbitmq/rabbitmq-amqp-go-client/pkg/rabbitmqamqp"
	"github.com/redis/go-redis/v9"
)

type GoBackServer struct {
	api.GoBackServer
	DB    *sqlx.DB
	Cache *redis.Client
	Queue *rmq.AmqpConnection
}

func NewGoBackServer(db *sqlx.DB, cache *redis.Client, queue *rmq.AmqpConnection) GoBackServer {
	gateway := GoBackServer{
		DB:    db,
		Cache: cache,
		Queue: queue,
	}

	return gateway
}

func (srv *GoBackServer) GetAccount(ctx context.Context, req *api.AccountRequest) (*api.AccountResponse, error) {
	slog.Info("Retrieving Account Information")
	accres := &api.AccountResponse{}
	return accres, nil
}

func (srv *GoBackServer) GetQuote(ctx context.Context, req *api.QuoteRequest) (*api.QuoteResponse, error) {
	slog.Info("Retrieving Stock Quote")
	qres := &api.QuoteResponse{}
	return qres, nil
}

func (srv *GoBackServer) GetTransactions(ctx context.Context, req *api.TransactionRequest) (*api.TransactionResponse, error) {
	slog.Info("Retrieving Transaction Logs")
	txLogs := []*api.Transaction{
		{TxId: uuid.NewV4().String(), UserId: uuid.NewV4().String(), Timestamp: time.Now().UTC().Format(time.RFC3339), Stock: "AAPL", Shares: 1000.0},
		{TxId: uuid.NewV4().String(), UserId: uuid.NewV4().String(), Timestamp: time.Now().UTC().Format(time.RFC3339), Stock: "GOOG", Shares: 5000.0},
	}

	txres := &api.TransactionResponse{
		Transactions: txLogs,
	}

	return txres, nil
}
