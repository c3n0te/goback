package main

import (
	"api"
	"fmt"
	"log/slog"

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
