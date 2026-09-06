package main

import (
	"api"
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"net"
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
	Quote net.Conn
}

func NewGoBackServer(db *sqlx.DB, cache *redis.Client, queue *rmq.AmqpConnection, quoteConn net.Conn) GoBackServer {
	gateway := GoBackServer{
		DB:    db,
		Cache: cache,
		Queue: queue,
		Quote: quoteConn,
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
	reader := bufio.NewReader(srv.Quote)
	srv.Quote.Write([]byte("AAPL\n"))
	resp, err := reader.ReadString('\n')
	if err != nil {
		slog.Error("Failed to read Quote Server quote: ", "error", err)
		return nil, err
	}

	slog.Info(fmt.Sprintf("Quote Server response: %v", resp))
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
