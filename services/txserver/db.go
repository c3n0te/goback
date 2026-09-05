package main

import (
	"github.com/jmoiron/sqlx"
)

func Migrate(db *sqlx.DB) {
	schema := `
	CREATE TABLE IF NOT EXISTS Accounts (
		userid			TEXT UNIQUE PRIMARY KEY,
      	username 		TEXT UNIQUE,
      	email 	 		TEXT UNIQUE,
        password  		TEXT,
        balance			DECIMAL
    );

    CREATE TABLE IF NOT EXISTS Transactions (
    	txid			TEXT UNIQUE PRIMARY KEY,
     	userid			TEXT,
        timestamp		TIMESTAMP,
    	stock			TEXT,
     	shares			DECIMAL
    );`

	db.MustExec(schema)
}
