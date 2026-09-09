package main

import (
	"api"
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"uuid"

	"github.com/jmoiron/sqlx"
	rmq "github.com/rabbitmq/rabbitmq-amqp-go-client/pkg/rabbitmqamqp"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GoBackServer struct {
	api.GoBackServer
	DB    *sqlx.DB
	Cache *redis.Client
	Queue *rmq.Consumer
	Quote net.Conn
}

func NewGoBackServer(db *sqlx.DB, cache *redis.Client, queue *rmq.Consumer, quoteConn net.Conn) GoBackServer {
	goback := GoBackServer{
		DB:    db,
		Cache: cache,
		Queue: queue,
		Quote: quoteConn,
	}

	return goback
}

func (srv *GoBackServer) HandleQueueMessage(ctx context.Context, delivery rmq.IDeliveryContext) {
	msg := delivery.Message()
	rawBytes := msg.Data[0]
	var newTxRequest api.TxRequest
	if err := json.Unmarshal(rawBytes, &newTxRequest); err != nil {
		slog.Error("Failed to unmarshal queue message into TxRequest: ", "error", err)
		return
	} else {
		delivery.Accept(ctx)
		slog.Info(fmt.Sprintf("TxRequest: %v", &newTxRequest))
	}

	transactionType := strings.ToUpper(newTxRequest.Type)
	switch transactionType {
	case "BUY":
		slog.Info("Buy Request")
	case "SELL":
		slog.Info("Sell Request")
	case "AUTOBUY":
		slog.Info("AutoBuy Request")
	case "AUTOSELL":
		slog.Info("AutoSell Request")
	default:
		slog.Error(fmt.Sprintf("Failed to match transaction type of TxRequest. Type: %v", transactionType))
		return
	}

	if err := InsertTransaction(srv.DB, &newTxRequest); err != nil {
		slog.Error("Failed to insert transaction: ", "error", err)
		return
	}
}

func (srv *GoBackServer) ConsumeQueue() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	for {
		delivery, err := srv.Queue.Receive(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				slog.Info("Shutting down Queue Consumer gracefully...")
				return
			}

			slog.Error("Failed to receive a message: ", "error", err)
			continue
		}

		go srv.HandleQueueMessage(ctx, delivery)
	}
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
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "UserId is required")
	}

	id := strings.TrimSpace(req.UserId)
	id = strings.ReplaceAll(id, "\x00", "")

	userId, err := uuid.Parse(id)
	if err != nil {
		slog.Error("Failed to parse AccountRequest UserId", "error", err)
		slog.Error(fmt.Sprintf("Failed UserId: %v", req.UserId))
		return nil, status.Errorf(codes.InvalidArgument, "Invalid ID format: %v", err)
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

func (srv *GoBackServer) GetTransactionLogs(ctx context.Context, req *api.TransactionLogRequest) (*api.TransactionLogResponse, error) {
	slog.Info("Retrieving Transaction Logs")
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "UserId is required")
	}

	id := strings.TrimSpace(req.UserId)
	id = strings.ReplaceAll(id, "\x00", "")

	userId, err := uuid.Parse(id)
	if err != nil {
		slog.Error("Failed to parse Transaction UserId", "error", err)
		slog.Error(fmt.Sprintf("Failed UserId: %v", req.UserId))
		return nil, status.Errorf(codes.InvalidArgument, "Invalid ID format: %v", err)
	}

	txLogs, err := ReadTransactionsWhereUserId(srv.DB, userId)
	if err != nil {
		slog.Error("Failed to retrieve transaction logs from DB", "error", err)
		return nil, err
	}

	txres := &api.TransactionLogResponse{
		UserId:          userId.String(),
		TransactionLogs: txLogs,
	}

	return txres, nil
}
