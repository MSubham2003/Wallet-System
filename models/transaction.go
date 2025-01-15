package models

import (
	"fmt"
	"time"
)

type Transaction struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Type      string    `json:"type"` // 'credit' or 'debit'
	Amount    float64   `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
}

// NewTransaction is a constructor to create a new transaction with the current time
func NewTransaction(userID int, transactionType string, amount float64) (*Transaction, error) {
	// Validate transaction type
	if transactionType != "credit" && transactionType != "debit" {
		return nil, fmt.Errorf("invalid transaction type: %s", transactionType)
	}
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}

	return &Transaction{
		UserID:    userID,
		Type:      transactionType,
		Amount:    amount,
		CreatedAt: time.Now(), // Set the transaction time to now
	}, nil
}
