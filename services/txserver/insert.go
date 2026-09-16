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
		slog.Error("Failed to create db transaction object: ", "error", err)
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

func InsertPortfolio(db *sqlx.DB, req *api.TxRequest, newShares float64) error {
	tx, err := db.Beginx()
	if err != nil {
		slog.Error("Failed to create db transaction object: ", "error", err)
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		`INSERT INTO Portfolios
			(userid, stock, shares)
		VALUES
			($1, $2, $3)
		ON CONFLICT (userid, stock)
		DO UPDATE SET shares = $4`,
		req.UserId,
		req.Stock,
		req.Shares,
		newShares,
	)

	if err != nil {
		slog.Error("Failed to insert transaction", "error", err)
		return err
	}

	tx.Commit()
	slog.Info("Portfolio row upserted")
	return nil
}

func InsertTransaction(db *sqlx.DB, req *api.TxRequest) error {
	tx, err := db.Beginx()
	if err != nil {
		slog.Error("Failed to create db transaction object: ", "error", err)
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		`INSERT INTO Transactions
			(txid, userid, type, timestamp, stock, shares, shareprice)
		VALUES
			($1, $2, $3, $4, $5, $6, $7)`,
		uuid.NewV4().String(),
		req.UserId,
		req.Type,
		time.Now().UTC().Format(time.RFC3339),
		req.Stock,
		req.Shares,
		req.SharePrice,
	)

	if err != nil {
		slog.Error("Failed to insert transaction", "error", err)
		return err
	}

	tx.Commit()
	slog.Info(fmt.Sprintf("Transaction inserted: %v", req))
	return nil
}
