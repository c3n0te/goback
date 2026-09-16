package api

import (
	"time"
	"uuid"
)

type Account struct {
	UserId   uuid.UUID `db:"userid"`
	Username string    `db:"username"`
	Email    string    `db:"email"`
	Password string    `db:"password"`
	Balance  float64   `db:"balance"`
}

type Transaction struct {
	TxId       uuid.UUID `db:"txid"`
	UserId     uuid.UUID `db:"userid"`
	Type       string    `db:"type"`
	Timestamp  time.Time `db:"timestamp"`
	Stock      string    `db:"stock"`
	Shares     float64   `db:"shares"`
	SharePrice float64   `db:"shareprice"`
}

type Portfolio struct {
	UserId uuid.UUID `db:"userid"`
	Stock  string    `db:"stock"`
	Shares float64   `db:"shares"`
}
