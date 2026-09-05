package main

import (
	"api"
	"context"
	"log/slog"

	"github.com/jmoiron/sqlx"
)

type GoBackServer struct {
	api.GoBackServer
	DB *sqlx.DB
}

func NewGatewayServer(db *sqlx.DB) GoBackServer {
	gateway := GoBackServer{
		DB: db,
	}

	return gateway
}

func (srv *GoBackServer) GetAccount(ctx context.Context, req *api.AccountRequest) (*api.AccountResponse, error) {
	slog.Info("Retrieving Account Information")
	accres := &api.AccountResponse{}
	return accres, nil
}

func (srv *GoBackServer) GetTransactions(ctx context.Context, req *api.TransactionRequest) (*api.TransactionResponse, error) {
	slog.Info("Retrieving Transaction Log")
	txres := &api.TransactionResponse{}
	return txres, nil
}
