package main

import (
	"api"
	"log/slog"
	"time"
	"uuid"

	"github.com/jmoiron/sqlx"
)

func ReadBalanceWhereUserId(db *sqlx.DB, userId uuid.UUID) (float32, error) {
	rows, err := db.Queryx(
		`SELECT
			balance
		FROM Accounts
		WHERE userid = $1
		LIMIT 1`,
		userId,
	)

	if err != nil {
		slog.Error("Failed to query Accounts table: ", "error", err)
		return 0.0, err
	}
	defer rows.Close()

	acc := api.Account{}
	for rows.Next() {
		err = rows.StructScan(&acc)
		if err != nil {
			slog.Error("Failed to marshal db rows into Account struct: ", "error", err)
			return 0.0, err
		}
	}

	return acc.Balance, nil
}

func ReadAccountWhereUserId(db *sqlx.DB, userId uuid.UUID) (*api.AccountResponse, error) {
	rows, err := db.Queryx(
		`SELECT
			userid,
			username,
			email,
			password,
			balance
		FROM Accounts
		WHERE userid = $1
		LIMIT 1`,
		userId,
	)

	if err != nil {
		slog.Error("Failed to query Accounts table: ", "error", err)
		return nil, err
	}
	defer rows.Close()

	acc := api.Account{}
	for rows.Next() {
		err = rows.StructScan(&acc)
		if err != nil {
			slog.Error("Failed to marshal db rows into Account struct: ", "error", err)
			return nil, err
		}
	}

	accres := &api.AccountResponse{
		UserId:   acc.UserId.String(),
		Username: acc.Username,
		Email:    acc.Email,
		Password: acc.Password,
		Balance:  acc.Balance,
	}

	return accres, nil
}

func ReadTransactionsWhereUserId(db *sqlx.DB, userId uuid.UUID) ([]*api.TransactionLog, error) {
	rows, err := db.Queryx(
		`SELECT
			txid,
			userid,
			type,
			timestamp,
			stock,
			shares
		FROM Transactions
		WHERE userid = $1`,
		userId,
	)

	if err != nil {
		slog.Error("Failed to query Transactions table: ", "error", err)
		return nil, err
	}
	defer rows.Close()

	txLogs := []*api.TransactionLog{}

	for rows.Next() {
		tx := api.Transaction{}
		err = rows.StructScan(&tx)
		if err != nil {
			slog.Error("Failed to marshal db rows into Transaction struct: ", "error", err)
			return nil, err
		}

		txLog := api.TransactionLog{
			TxId:      tx.TxId.String(),
			UserId:    tx.UserId.String(),
			Type:      tx.Type,
			Timestamp: tx.Timestamp.Format(time.RFC3339),
			Stock:     tx.Stock,
			Shares:    tx.Shares,
		}

		txLogs = append(txLogs, &txLog)
	}

	return txLogs, nil
}
