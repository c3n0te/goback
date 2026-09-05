package main

import (
	"api"
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc/metadata"
)

func CallGetAccount(gbc api.GoBackClient) (*api.AccountResponse, error) {
	md := metadata.Pairs("timestamp", time.Now().UTC().Format(time.StampNano))
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	accreq := &api.AccountRequest{}

	resp, err := gbc.GetAccount(ctx, accreq)
	if err != nil {
		slog.Error("Failed to register gateway gRPC client", "error", err)
		return nil, err
	}

	return resp, nil
}

func CallGetTransactions(gbc api.GoBackClient) (*api.TransactionResponse, error) {
	md := metadata.Pairs("timestamp", time.Now().UTC().Format(time.StampNano))
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	txreq := &api.TransactionRequest{}

	resp, err := gbc.GetTransactions(ctx, txreq)
	if err != nil {
		slog.Error("Failed to register gateway gRPC client", "error", err)
		return nil, err
	}

	return resp, nil
}
