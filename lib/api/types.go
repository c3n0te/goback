package api

import (
	"time"
	"uuid"
)

type DbTransaction struct {
	TxId      uuid.UUID `db:"txid"`
	UserId    uuid.UUID `db:"userid"`
	Type      string    `db:"type"`
	Timestamp time.Time `db:"timestamp"`
	Stock     string    `db:"stock"`
	Shares    float32   `db:"shares"`
}
