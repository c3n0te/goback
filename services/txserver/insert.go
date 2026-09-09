package main

import (
	"api"
	"fmt"
	"log/slog"
	"time"
	"uuid"

	"github.com/jmoiron/sqlx"
)

func InsertAccount(db *sqlx.DB, req *api.CreateAccountRequest) (*api.CreateAccountResponse, error) {
	tx, err := db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		`INSERT INTO Accounts
			(userid, username, email, password, balance)
		VALUES
			($1, $2, $3, $4, $5)`,
		req.UserId,
		req.Username,
		req.Email,
		req.Password,
		req.Balance,
	)

	if err != nil {
		slog.Error("Failed to insert user account", "error", err)
		return nil, err
	}

	tx.Commit()
	slog.Info(fmt.Sprintf("Account inserted: %v", req))
	createAccountRes := &api.CreateAccountResponse{
		UserId: req.UserId,
	}

	return createAccountRes, nil
}

func InsertTransaction(db *sqlx.DB, req *api.TxRequest) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	ts := time.Now().UTC().Format(time.RFC3339)
	_, err = tx.Exec(
		`INSERT INTO Transactions
			(txid, userid, type, timestamp, stock, shares)
		VALUES
			($1, $2, $3, $4, $5, $6)`,
		uuid.NewV4().String(),
		req.UserId,
		req.Type,
		ts,
		req.Stock,
		req.Shares,
	)

	if err != nil {
		slog.Error("Failed to insert transaction", "error", err)
		return err
	}

	tx.Commit()
	slog.Info(fmt.Sprintf("Transaction inserted: %v", req))
	return nil
}
