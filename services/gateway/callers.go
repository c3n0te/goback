package main

import (
	"api"
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc/metadata"
)

func CallGetAccount(gbc api.GoBackClient, userId string) (*api.AccountResponse, error) {
	md := metadata.Pairs("timestamp", time.Now().UTC().Format(time.StampNano))
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	accreq := &api.AccountRequest{
		UserId: userId,
	}

	resp, err := gbc.GetAccount(ctx, accreq)
	if err != nil {
		slog.Error("Failed to register gateway gRPC client: ", "error", err)
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

	resp, err := gbc.GetQuote(ctx, qreq)
	if err != nil {
		slog.Error("Failed to retrieve stock quote: ", "error", err)
		return nil, err
	}

	return resp, nil
}

func CallGetTransactions(gbc api.GoBackClient, userId string) (*api.TransactionResponse, error) {
	md := metadata.Pairs("timestamp", time.Now().UTC().Format(time.StampNano))
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	txreq := &api.TransactionRequest{
		UserId: userId,
	}

	resp, err := gbc.GetTransactions(ctx, txreq)
	if err != nil {
		slog.Error("Failed to retrieve transaction logs: ", "error", err)
		return nil, err
	}

	return resp, nil
}
