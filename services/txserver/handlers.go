package main

import (
	"api"
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
	"uuid"

	amqp "github.com/Azure/go-amqp"
	"github.com/jmoiron/sqlx"
	rmq "github.com/rabbitmq/rabbitmq-amqp-go-client/pkg/rabbitmqamqp"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AutoTxResult struct {
	SharePrice float64
	Error      error
}

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
		err = delivery.Discard(ctx, &amqp.Error{
			Condition:   amqp.ErrCondInvalidField,
			Description: err.Error(),
		})

		if err != nil {
			slog.Error("Failed to communicate delivery error back to Queue", "error", err)
		}

		return
	}

	slog.Info(fmt.Sprintf("TxRequest: %v", &newTxRequest))
	userId, err := uuid.Parse(newTxRequest.UserId)
	if err != nil {
		slog.Error("Failed to parse UUID: ", "error", err)
		slog.Error(fmt.Sprintf("Failed UUID: %v", newTxRequest.UserId))
		err = delivery.Discard(ctx, &amqp.Error{
			Condition:   amqp.ErrCondInvalidField,
			Description: err.Error(),
		})

		if err != nil {
			slog.Error("Failed to communicate delivery error back to Queue", "error", err)
		}

		return
	}

	currShares, err := ReadPortfolioSharesWhereUserIdAndStock(srv.DB, userId, newTxRequest.Stock)
	if err != nil {
		slog.Error("Failed to read current portfolio shares: ", "error", err)
		delivery.Requeue(ctx)
		return
	}

	currBalance, err := ReadBalanceWhereUserId(srv.DB, userId)
	if err != nil {
		slog.Error("Failed to read current portfolio shares: ", "error", err)
		delivery.Requeue(ctx)
		return
	}

	newShares, newBalance := srv.HandleTxRequestType(&newTxRequest, currShares, currBalance)
	if math.Signbit(newShares) || math.Signbit(newBalance) {
		slog.Error("Invalid operation, cannot have negative balance or shares")
		err = delivery.Discard(ctx, &amqp.Error{
			Condition:   amqp.ErrCondIllegalState,
			Description: "Invalid Operation, cannot have negative balance or shares",
		})

		if err != nil {
			slog.Error("Failed to communicate delivery error back to Queue", "error", err)
		}

		return
	}

	if _, err := UpdateBalance(srv.DB, userId, newBalance); err != nil {
		slog.Error("Failed to update balance", "error", err)
		delivery.Requeue(ctx)
		return
	}

	if err := InsertPortfolioOnConflictNewShares(srv.DB, &newTxRequest, newShares); err != nil {
		slog.Error("Failed to insert portfolio row", "error", err)
		delivery.Requeue(ctx)
		return
	}

	if err := InsertTransaction(srv.DB, &newTxRequest); err != nil {
		slog.Error("Failed to insert transaction: ", "error", err)
		delivery.Requeue(ctx)
		return
	}

	err = delivery.Accept(ctx)
	if err != nil {
		slog.Error("Failed to communicate delivery acceptance back to Queue", "error", err)
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

func (srv *GoBackServer) HandleAutoBuyRequest(ch chan *AutoTxResult, stock string, tgtPrice float64) {
	currStockPriceFloat := tgtPrice
	var currStockPrice string

	for currStockPriceFloat >= tgtPrice {
		reader := bufio.NewReader(srv.Quote)
		msg := fmt.Sprintf("%v\n", stock)
		srv.Quote.Write([]byte(msg))
		resp, err := reader.ReadString('\n')
		if err != nil {
			slog.Error("Failed to read Quote Server quote: ", "error", err)
			res := &AutoTxResult{
				SharePrice: -1.0,
				Error:      err,
			}

			ch <- res
			return
		}

		slog.Info(fmt.Sprintf("Quote Server response: %v", resp))
		currStockPrice = strings.Split(resp, ",")[1]

		currStockPriceFloat, err = strconv.ParseFloat(currStockPrice, 64)
		if err != nil {
			slog.Error("Failed to convert stock price string to float", "error", err)
			res := &AutoTxResult{
				SharePrice: -1.0,
				Error:      err,
			}

			ch <- res
			return
		}
	}

	res := &AutoTxResult{
		SharePrice: currStockPriceFloat,
		Error:      nil,
	}

	ch <- res
}

func (srv *GoBackServer) HandleAutoSellRequest(ch chan *AutoTxResult, stock string, tgtPrice float64) {
	currStockPriceFloat := tgtPrice
	var currStockPrice string

	for currStockPriceFloat <= tgtPrice {
		reader := bufio.NewReader(srv.Quote)
		msg := fmt.Sprintf("%v\n", stock)
		srv.Quote.Write([]byte(msg))
		resp, err := reader.ReadString('\n')
		if err != nil {
			slog.Error("Failed to read Quote Server quote: ", "error", err)
			res := &AutoTxResult{
				SharePrice: -1.0,
				Error:      err,
			}

			ch <- res
			return
		}

		slog.Info(fmt.Sprintf("Quote Server response: %v", resp))
		currStockPrice = strings.Split(resp, ",")[1]

		currStockPriceFloat, err = strconv.ParseFloat(currStockPrice, 64)
		if err != nil {
			slog.Error("Failed to convert stock price string to float", "error", err)
			res := &AutoTxResult{
				SharePrice: -1.0,
				Error:      err,
			}

			ch <- res
			return
		}
	}

	res := &AutoTxResult{
		SharePrice: currStockPriceFloat,
		Error:      nil,
	}

	ch <- res
}

func (srv *GoBackServer) HandleTxRequestType(newTxRequest *api.TxRequest, currShares float64, currBalance float64) (float64, float64) {
	newShares := float64(0.0)
	newBalance := float64(0.0)
	transactionType := strings.ToUpper(newTxRequest.Type)

	switch transactionType {
	case "BUY":
		newShares = currShares + newTxRequest.Shares
		newBalance = currBalance - (newTxRequest.Shares * newTxRequest.SharePrice)

	case "SELL":
		newShares = currShares - newTxRequest.Shares
		newBalance = currBalance + (newTxRequest.Shares * newTxRequest.SharePrice)

	case "AUTOBUY":
		ch := make(chan *AutoTxResult)
		go srv.HandleAutoBuyRequest(ch, newTxRequest.Stock, newTxRequest.SharePrice)
		txRes := <-ch

		if txRes.Error != nil {
			slog.Error("Failed to handle AutoBuy request", "error", txRes.Error)
			return -1.0, -1.0
		}

		slog.Info(fmt.Sprintf("Updating TxRequest with new Quoted SharePrice, old price: %v; new price: %v", newTxRequest.SharePrice, txRes.SharePrice))
		newTxRequest.SharePrice = txRes.SharePrice
		newShares = currShares + newTxRequest.Shares
		newBalance = currBalance - (newTxRequest.Shares * newTxRequest.SharePrice)

	case "AUTOSELL":
		ch := make(chan *AutoTxResult)
		go srv.HandleAutoSellRequest(ch, newTxRequest.Stock, newTxRequest.SharePrice)
		txRes := <-ch

		if txRes.Error != nil {
			slog.Error("Failed to handle AutoSell request", "error", txRes.Error)
			return -1.0, -1.0
		}

		slog.Info(fmt.Sprintf("Updating TxRequest with new Quoted SharePrice, old price: %v; new price: %v", newTxRequest.SharePrice, txRes.SharePrice))
		newTxRequest.SharePrice = txRes.SharePrice
		newShares = currShares - newTxRequest.Shares
		newBalance = currBalance + (newTxRequest.Shares * newTxRequest.SharePrice)

	default:
		slog.Error(fmt.Sprintf("Failed to match transaction type of TxRequest. Type: %v", transactionType))
		return -1.0, -1.0
	}

	return newShares, newBalance
}

func (srv *GoBackServer) GetPortfolio(ctx context.Context, req *api.PortfolioRequest) (*api.PortfolioResponse, error) {
	slog.Info("Retrieving Portfolio")
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

	pLogs, err := ReadPortfolioWhereUserId(srv.DB, userId)
	if err != nil {
		slog.Error("Failed to retrieve transaction logs from DB", "error", err)
		return nil, err
	}

	pres := &api.PortfolioResponse{
		UserId:        userId.String(),
		PortfolioLogs: pLogs,
	}

	return pres, nil
}

func (srv *GoBackServer) AddBalance(ctx context.Context, req *api.BalanceRequest) (*api.BalanceResponse, error) {
	slog.Info("Adding Balance to existing account")
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		slog.Error("Failed to parse UUID: ", "error", err)
		slog.Error(fmt.Sprintf("Failed UUID: %v", userId))
		return nil, err
	}

	currBalance, err := ReadBalanceWhereUserId(srv.DB, userId)
	if err != nil {
		slog.Error("Failed to read user balance: ", "error", err)
		return nil, err
	}

	newBalance := currBalance + req.AddAmount
	addBalanceRes, err := UpdateBalance(srv.DB, userId, newBalance)
	if err != nil {
		slog.Error("Failed to upsert balance amount to existing account", "error", err)
		return nil, err
	}

	return addBalanceRes, nil
}

func (srv *GoBackServer) CreateAccount(ctx context.Context, req *api.CreateAccountRequest) (*api.CreateAccountResponse, error) {
	slog.Info("Creating New User Account")
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
