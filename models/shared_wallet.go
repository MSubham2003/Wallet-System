package models

import "time"

type SharedWallet struct {
	Balance     float64        `json:"balance"` // Shared balance
	CreatedAt   time.Time      `json:"created_at"`
	Transactions []Transaction `json:"transactions"` // List of transactions on the shared wallet
	// Locked      bool           `json:"locked"` // Indicates whether the wallet is locked for a transaction
}
