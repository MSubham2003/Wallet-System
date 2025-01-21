package models

import "time"

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
	Fname     string    `json:"fname"`
	Lname     string    `json:"lname"`
	Email     string    `json:"email"`
}
