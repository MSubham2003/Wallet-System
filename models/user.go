package models

import "time"

type User struct {
	ID           int           `json:"id"`
	Username     string        `json:"username"`
	Balance      float64       `json:"balance"`
	CreatedAt    time.Time     `json:"created_at"`
	Transactions []Transaction `json:"transactions"` // Foreign key relation
}
