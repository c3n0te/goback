package main

import (
	"api"
	"fmt"
	"log/slog"
	"strings"
)

func calcNewSharesAndBalance(newTxRequest *api.TxRequest, currShares float64, currBalance float64) (float64, float64) {
	newShares := float64(0.0)
	newBalance := float64(0.0)
	transactionType := strings.ToUpper(newTxRequest.Type)

	switch transactionType {
	case "BUY":
		newShares = currShares + newTxRequest.Shares
		newBalance = currBalance - (newTxRequest.Shares * newTxRequest.SharePrice)

	case "SELL":
		newShares = currShares - newTxRequest.Shares
		newBalance = currBalance + (newTxRequest.Shares * newTxRequest.SharePrice)

	case "AUTOBUY":
		newShares = currShares + newTxRequest.Shares
		newBalance = currBalance - (newTxRequest.Shares * newTxRequest.SharePrice)

	case "AUTOSELL":
		newShares = currShares - newTxRequest.Shares
		newBalance = currBalance + (newTxRequest.Shares * newTxRequest.SharePrice)

	default:
		slog.Error(fmt.Sprintf("Failed to match transaction type of TxRequest. Type: %v", transactionType))
		return -1.0, -1.0
	}

	return newShares, newBalance
}
