package models

import "time"

type Transaction struct {
	TransactionID        int       `json:"id"`         // Unique transaction ID
	UserID    int       `json:"user_id"`    // User who initiated the transaction
	UserName  string    `json:"user_name"`  // UserName who initiated the transaction
	Type      string    `json:"type"`       // Transaction type (e.g., "debit" or "credit")
	Amount    float64   `json:"amount"`     // Transaction amount
	CreatedAt time.Time `json:"created_at"` // Timestamp of when the transaction occurred
	Status    string    `json:"status"`     // Status of the transaction (e.g., "completed", "pending", "failed")
}
