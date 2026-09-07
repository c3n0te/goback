package main

import (
	"api"
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"net"
	"strings"
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
	stockPrice := ""
	var err error

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	stockPrice, err = srv.Cache.Get(ctx, req.Stock).Result()
	if err == redis.Nil {
		slog.Info("Stock Price Not In Cache")
		reader := bufio.NewReader(srv.Quote)
		msg := fmt.Sprintf("%v\n", req.Stock)
		srv.Quote.Write([]byte(msg))
		resp, err := reader.ReadString('\n')
		if err != nil {
			slog.Error("Failed to read Quote Server quote: ", "error", err)
			return nil, err
		}

		slog.Info(fmt.Sprintf("Quote Server response: %v", resp))
		stockPrice = strings.Split(resp, ",")[1]
		err = srv.Cache.Set(ctx, req.Stock, stockPrice, 1*time.Second).Err()
		if err != nil {
			slog.Error("Failed to set stock price in cache", "error", err)
			return nil, err
		}
	} else if err != nil {
		slog.Error("Failed to retrieve stock ticker from cache", "error", err)
		return nil, err
	}

	qres := &api.QuoteResponse{
		Stock: req.Stock,
		Price: stockPrice,
	}

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
