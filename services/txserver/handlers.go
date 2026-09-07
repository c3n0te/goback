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
	goback := GoBackServer{
		DB:    db,
		Cache: cache,
		Queue: queue,
		Quote: quoteConn,
	}

	return goback
}

func (srv *GoBackServer) CreateAccount(ctx context.Context, req *api.CreateAccountRequest) (*api.CreateAccountResponse, error) {
	slog.Info("Creating New User Account")
	req.UserId = uuid.NewV4().String()
	createAccountRes, err := InsertAccount(srv.DB, req)
	if err != nil {
		slog.Error("Failed to insert account information", "error", err)
		return nil, err
	}

	return createAccountRes, nil
}

func (srv *GoBackServer) GetAccount(ctx context.Context, req *api.AccountRequest) (*api.AccountResponse, error) {
	slog.Info("Retrieving Account Information")
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		slog.Error("Failed to parse Transaction UserId", "error", err)
		return nil, err
	}

	accres, err := ReadAccountWhereUserId(srv.DB, userId)
	if err != nil {
		slog.Error("Failed to retrieve account information from DB", "error", err)
		return nil, err
	}

	return accres, nil
}

func (srv *GoBackServer) GetQuote(ctx context.Context, req *api.QuoteRequest) (*api.QuoteResponse, error) {
	slog.Info("Retrieving Stock Quote")
	var stockPrice string
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
	} else {
		slog.Info("Stock Price In Cache")
	}

	qres := &api.QuoteResponse{
		Stock: req.Stock,
		Price: stockPrice,
	}

	return qres, nil
}

func (srv *GoBackServer) GetTransactions(ctx context.Context, req *api.TransactionRequest) (*api.TransactionResponse, error) {
	slog.Info("Retrieving Transaction Logs")
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		slog.Error("Failed to parse Transaction UserId", "error", err)
		return nil, err
	}

	txLogs, err := ReadTransactions(srv.DB, userId)
	if err != nil {
		slog.Error("Failed to retrieve transaction logs from DB", "error", err)
		return nil, err
	}

	txres := &api.TransactionResponse{
		UserId:       userId.String(),
		Transactions: txLogs,
	}

	return txres, nil
}
