package main

import (
	"api"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	rmq "github.com/rabbitmq/rabbitmq-amqp-go-client/pkg/rabbitmqamqp"
	"google.golang.org/grpc/metadata"
)

func CallQueuePublishSell(qpublisher *rmq.Publisher, newSellReq *api.TxRequest) (*api.TxResponse, error) {
	slog.Info("Publishing SellRequest to Message Queue")
	body, err := json.Marshal(newSellReq)
	if err != nil {
		slog.Error("Failed to marshal BuyRequest", "error", err)
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resp, err := qpublisher.Publish(ctx, rmq.NewMessage([]byte(body)))
	switch resp.Outcome.(type) {
	case *rmq.StateAccepted:
		sellResp := &api.TxResponse{
			Status:    true,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}

		return sellResp, nil

	case *rmq.StateRejected:
		slog.Error(fmt.Sprintf("Message was rejected: %v", resp.Outcome))
		return nil, err

	case *rmq.StateReleased:
		slog.Error(fmt.Sprintf("Message was released: %v", resp.Outcome))
		return nil, err

	case *rmq.StateModified:
		slog.Error(fmt.Sprintf("Message was modified: %v", resp.Outcome))
		return nil, err

	default:
		slog.Error(fmt.Sprintf("Unexpected publish outcome: %v", resp.Outcome))
		return nil, err
	}
}

func CallQueuePublishBuy(qpublisher *rmq.Publisher, newBuyReq *api.TxRequest) (*api.TxResponse, error) {
	slog.Info("Publishing BuyRequest to Message Queue")
	body, err := json.Marshal(newBuyReq)
	if err != nil {
		slog.Error("Failed to marshal BuyRequest", "error", err)
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resp, err := qpublisher.Publish(ctx, rmq.NewMessage([]byte(body)))
	switch resp.Outcome.(type) {
	case *rmq.StateAccepted:
		buyResp := &api.TxResponse{
			Status:    true,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}

		return buyResp, nil

	case *rmq.StateRejected:
		slog.Error(fmt.Sprintf("Message was rejected: %v", resp.Outcome))
		return nil, err

	case *rmq.StateReleased:
		slog.Error(fmt.Sprintf("Message was released: %v", resp.Outcome))
		return nil, err

	case *rmq.StateModified:
		slog.Error(fmt.Sprintf("Message was modified: %v", resp.Outcome))
		return nil, err

	default:
		slog.Error(fmt.Sprintf("Unexpected publish outcome: %v", resp.Outcome))
		return nil, err
	}
}

func CallCreateAccount(gbc api.GoBackClient, newAccount *api.CreateAccountRequest) (*api.CreateAccountResponse, error) {
	md := metadata.Pairs("timestamp", time.Now().UTC().Format(time.StampNano))
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	slog.Info("Calling CreateAccount")
	resp, err := gbc.CreateAccount(ctx, newAccount)
	if err != nil {
		slog.Error("Failed to create account: ", "error", err)
		return nil, err
	}

	return resp, nil
}

func CallGetAccount(gbc api.GoBackClient, userId string) (*api.AccountResponse, error) {
	md := metadata.Pairs("timestamp", time.Now().UTC().Format(time.StampNano))
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	accreq := &api.AccountRequest{
		UserId: userId,
	}

	slog.Info("Calling GetAccount")
	resp, err := gbc.GetAccount(ctx, accreq)
	if err != nil {
		slog.Error("Failed to get account information: ", "error", err)
		return nil, err
	}

	return resp, nil
}

func CallGetQuote(gbc api.GoBackClient, ticker string) (*api.QuoteResponse, error) {
	md := metadata.Pairs("timestamp", time.Now().UTC().Format(time.StampNano))
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	qreq := &api.QuoteRequest{
		Stock: ticker,
	}

	slog.Info("Calling GetQuote")
	resp, err := gbc.GetQuote(ctx, qreq)
	if err != nil {
		slog.Error("Failed to retrieve stock quote: ", "error", err)
		return nil, err
	}

	return resp, nil
}

func CallGetTransactionLogs(gbc api.GoBackClient, userId string) (*api.TransactionLogResponse, error) {
	md := metadata.Pairs("timestamp", time.Now().UTC().Format(time.StampNano))
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	txreq := &api.TransactionLogRequest{
		UserId: userId,
	}

	slog.Info("Calling GetTransactions")
	resp, err := gbc.GetTransactionLogs(ctx, txreq)
	if err != nil {
		slog.Error("Failed to retrieve transaction logs: ", "error", err)
		return nil, err
	}

	return resp, nil
}
