package main

import (
	"api"
	"context"
	"log/slog"
	"time"
	"uuid"

	"google.golang.org/grpc/metadata"
)

func CallCreateAccount(gbc api.GoBackClient) (*api.CreateAccountResponse, error) {
	md := metadata.Pairs("timestamp", time.Now().UTC().Format(time.StampNano))
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	createAccountReq := &api.CreateAccountRequest{
		UserId:   uuid.NewV4().String(),
		Username: "c3n0te",
		Email:    "c3n0te@gmail.com",
		Password: "secret",
		Balance:  1000.0,
	}

	slog.Info("Calling CreateAccount")
	resp, err := gbc.CreateAccount(ctx, createAccountReq)
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

func CallGetTransactions(gbc api.GoBackClient, userId string) (*api.TransactionResponse, error) {
	md := metadata.Pairs("timestamp", time.Now().UTC().Format(time.StampNano))
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	txreq := &api.TransactionRequest{
		UserId: userId,
	}

	slog.Info("Calling GetTransactions")
	resp, err := gbc.GetTransactions(ctx, txreq)
	if err != nil {
		slog.Error("Failed to retrieve transaction logs: ", "error", err)
		return nil, err
	}

	return resp, nil
}
