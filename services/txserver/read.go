package main

import (
	"api"
	"log/slog"
	"time"
	"uuid"

	"github.com/jmoiron/sqlx"
)

func ReadPortfolioSharesWhereUserIdAndStock(db *sqlx.DB, userId uuid.UUID, stock string) (float64, error) {
	rows, err := db.Queryx(
		`SELECT
			shares
		FROM Portfolios
		WHERE userid = $1
		AND stock = $2
		LIMIT 1`,
		userId,
		stock,
	)

	if err != nil {
		slog.Error("Failed to query Portfolios table: ", "error", err)
		return 0.0, err
	}
	defer rows.Close()

	portf := api.Portfolio{}
	for rows.Next() {
		err = rows.StructScan(&portf)
		if err != nil {
			slog.Error("Failed to marshal db rows into Portfolio struct: ", "error", err)
			return 0.0, err
		}
	}

	return portf.Shares, nil
}

func ReadPortfolioWhereUserId(db *sqlx.DB, userId uuid.UUID) ([]*api.PortfolioLog, error) {
	rows, err := db.Queryx(
		`SELECT
			stock,
			shares
		FROM Portfolios
		WHERE userid = $1`,
		userId,
	)

	if err != nil {
		slog.Error("Failed to query Portfolios table: ", "error", err)
		return nil, err
	}
	defer rows.Close()

	pLogs := []*api.PortfolioLog{}

	for rows.Next() {
		portf := api.Portfolio{}
		err = rows.StructScan(&portf)
		if err != nil {
			slog.Error("Failed to marshal db rows into Portfolio struct: ", "error", err)
			return nil, err
		}

		pLog := api.PortfolioLog{
			Stock:  portf.Stock,
			Shares: portf.Shares,
		}

		pLogs = append(pLogs, &pLog)
	}

	return pLogs, nil
}

func ReadBalanceWhereUserId(db *sqlx.DB, userId uuid.UUID) (float64, error) {
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
			TxId:       tx.TxId.String(),
			Type:       tx.Type,
			Timestamp:  tx.Timestamp.Format(time.RFC3339),
			Stock:      tx.Stock,
			Shares:     tx.Shares,
			SharePrice: tx.SharePrice,
		}

		txLogs = append(txLogs, &txLog)
	}

	return txLogs, nil
}
