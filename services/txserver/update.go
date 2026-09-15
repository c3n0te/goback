package main

import (
	"api"
	"log/slog"
	"uuid"

	"github.com/jmoiron/sqlx"
)

func UpdateBalance(db *sqlx.DB, userId uuid.UUID, addAmount float32, currBalance float32) (*api.BalanceResponse, error) {
	tx, err := db.Beginx()
	if err != nil {
		slog.Error("Failed to create db transaction object: ", "error", err)
		return nil, err
	}
	defer tx.Rollback()

	newBalance := addAmount + currBalance
	_, err = tx.Exec(
		`UPDATE Accounts SET balance = $1 WHERE userid = $2`,
		newBalance,
		userId,
	)

	if err != nil {
		slog.Error("Failed to update account balance", "error", err)
		return nil, err
	}

	tx.Commit()
	bres := &api.BalanceResponse{
		Status:     true,
		NewBalance: newBalance,
	}

	return bres, nil
}

func UpdatePortfolio(db *sqlx.DB, userId uuid.UUID, stock string, addShares float32, currShares float32) error {
	tx, err := db.Beginx()
	if err != nil {
		slog.Error("Failed to create db transaction object: ", "error", err)
		return err
	}
	defer tx.Rollback()

	newShares := currShares + addShares
	_, err = tx.Exec(
		`UPDATE Portfolios SET shares = $1 WHERE userid = $2 AND stock = $3`,
		newShares,
		userId,
		stock,
	)

	if err != nil {
		slog.Error("Failed to update portfolio", "error", err)
		return err
	}

	tx.Commit()
	return nil
}
