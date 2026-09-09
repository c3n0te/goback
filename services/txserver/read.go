package main

import (
	"api"
	"log/slog"
	"uuid"

	"github.com/jmoiron/sqlx"
)

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

	accres := &api.AccountResponse{}
	for rows.Next() {
		err = rows.StructScan(&accres)
		if err != nil {
			slog.Error("Failed to marshal db rows into AccountResponse struct: ", "error", err)
			return nil, err
		}
	}

	return accres, nil
}

func ReadTransactionsWhereUserId(db *sqlx.DB, userId uuid.UUID) ([]*api.Transaction, error) {
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

	txs := []*api.Transaction{}

	for rows.Next() {
		dbtx := api.DbTransaction{}
		err = rows.StructScan(&dbtx)
		if err != nil {
			slog.Error("Failed to marshal db rows into DbTransaction struct: ", "error", err)
			return nil, err
		}

		tx := api.Transaction{
			TxId:      dbtx.TxId.String(),
			UserId:    dbtx.UserId.String(),
			Type:      dbtx.Type,
			Timestamp: dbtx.Timestamp.String(),
			Stock:     dbtx.Stock,
			Shares:    dbtx.Shares,
		}

		txs = append(txs, &tx)
	}

	return txs, nil
}
