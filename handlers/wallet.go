package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"wallet-system/models"

	"github.com/gorilla/mux"
)

func CreateUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the JSON body
		var requestData struct {
			Username string `json:"username"`
		}

		// Decode JSON
		err := json.NewDecoder(r.Body).Decode(&requestData)
		if err != nil {
			http.Error(w, "Invalid JSON input", http.StatusBadRequest)
			return
		}

		// Validate input
		if requestData.Username == "" {
			http.Error(w, "username is required", http.StatusBadRequest)
			return
		}

		// Insert user into the database
		query := `INSERT INTO users (username, balance) VALUES ($1, 0) RETURNING id`
		var id int
		err = db.QueryRow(query, requestData.Username).Scan(&id)
		if err != nil {
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			fmt.Print(err)
			return
		}

		// Respond with success
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "User created successfully with ID: %d\n", id)
	}
}

// UpdateUser updates a user's details (e.g., username)
func UpdateUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user ID from URL
		id := mux.Vars(r)["id"]
		if id == "" {
			http.Error(w, "id is required", http.StatusBadRequest)
			return
		}

		// Get the updated username from the request body
		var user models.User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		// Update the user in the database
		query := `UPDATE users SET username = $1 WHERE id = $2`
		_, err := db.Exec(query, user.Username, id)
		if err != nil {
			http.Error(w, "Failed to update user", http.StatusInternalServerError)
			return
		}

		// Send a success response
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "User with ID %s updated successfully", id)
	}
}

// DeleteUser deletes a user from the database
func DeleteUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract the `id` from the path parameters
		vars := mux.Vars(r)
		id := vars["id"] // Path parameter `id`

		if id == "" {
			http.Error(w, "id is required", http.StatusBadRequest)
			return
		}

		// Check if the user exists
		var exists bool
		checkQuery := `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`
		err := db.QueryRow(checkQuery, id).Scan(&exists)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Failed to check user existence", http.StatusInternalServerError)
			return
		}

		if !exists {
			http.Error(w, "No user found with the provided id", http.StatusNotFound)
			return
		}

		// Execute the DELETE query
		deleteQuery := `DELETE FROM users WHERE id = $1`
		_, err = db.Exec(deleteQuery, id)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Failed to delete user", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "User deleted successfully")
	}
}

// CreditBalance credits an amount to a user's wallet
func CreditBalance(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the JSON body
		var requestData struct {
			ID     string  `json:"id"`
			Amount float64 `json:"amount"`
		}
		err := json.NewDecoder(r.Body).Decode(&requestData)
		if err != nil {
			http.Error(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}

		// Validate the input
		if requestData.ID == "" || requestData.Amount <= 0 {
			http.Error(w, "id and valid amount are required", http.StatusBadRequest)
			return
		}

		// Begin a transaction
		tx, err := db.Begin()
		if err != nil {
			http.Error(w, "Failed to begin transaction", http.StatusInternalServerError)
			return
		}

		// Update the user's balance
		_, err = tx.Exec(`UPDATE users SET balance = balance + $1 WHERE id = $2`, requestData.Amount, requestData.ID)
		if err != nil {
			tx.Rollback()
			http.Error(w, "Failed to credit balance", http.StatusInternalServerError)
			return
		}

		// Log the transaction
		_, err = tx.Exec(`INSERT INTO transactions (user_id, type, amount) VALUES ($1, 'credit', $2)`, requestData.ID, requestData.Amount)
		if err != nil {
			tx.Rollback()
			http.Error(w, "Failed to log transaction", http.StatusInternalServerError)
			return
		}

		// Commit the transaction
		tx.Commit()
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Balance credited successfully")
	}
}

// DebitBalance debits an amount from a user's wallet
func DebitBalance(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the JSON body
		var requestData struct {
			ID     string  `json:"id"`
			Amount float64 `json:"amount"`
		}
		err := json.NewDecoder(r.Body).Decode(&requestData)
		if err != nil {
			http.Error(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}

		// Validate input
		if requestData.ID == "" || requestData.Amount <= 0 {
			http.Error(w, "id and valid amount are required", http.StatusBadRequest)
			return
		}

		// Begin a transaction
		tx, err := db.Begin()
		if err != nil {
			http.Error(w, "Failed to begin transaction", http.StatusInternalServerError)
			return
		}

		// Check user's balance
		var balance float64
		err = tx.QueryRow(`SELECT balance FROM users WHERE id = $1`, requestData.ID).Scan(&balance)
		if err != nil {
			tx.Rollback()
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		if balance < requestData.Amount {
			tx.Rollback()
			http.Error(w, "Insufficient balance", http.StatusBadRequest)
			return
		}

		// Update user's balance
		_, err = tx.Exec(`UPDATE users SET balance = balance - $1 WHERE id = $2`, requestData.Amount, requestData.ID)
		if err != nil {
			tx.Rollback()
			http.Error(w, "Failed to debit balance", http.StatusInternalServerError)
			return
		}

		// Log the transaction
		_, err = tx.Exec(`INSERT INTO transactions (user_id, type, amount) VALUES ($1, 'debit', $2)`, requestData.ID, requestData.Amount)
		if err != nil {
			tx.Rollback()
			http.Error(w, "Failed to log transaction", http.StatusInternalServerError)
			return
		}

		// Commit the transaction
		tx.Commit()
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Balance debited successfully")
	}
}

// GetTransactions retrieves all transactions for a user
func GetTransactions(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract ID from the path parameters
		vars := mux.Vars(r)
		id := vars["id"]
		if id == "" {
			http.Error(w, "id is required", http.StatusBadRequest)
			return
		}

		// Query to get transactions along with the user's username
		rows, err := db.Query(`
			SELECT u.id, u.username, t.id, t.type, t.amount, t.created_at 
			FROM transactions t
			JOIN users u ON t.user_id = u.id
			WHERE t.user_id = $1
			ORDER BY t.created_at DESC`, id)
		if err != nil {
			http.Error(w, "Failed to fetch transactions", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var userInfo struct {
			UserID   int    `json:"user_id"`
			Username string `json:"username"`
		}

		// Initialize the response
		var transactions []map[string]interface{}
		for rows.Next() {
			var t struct {
				ID        int       `json:"id"`
				Type      string    `json:"type"`
				Amount    float64   `json:"amount"`
				CreatedAt time.Time `json:"created_at"`
			}

			err := rows.Scan(&userInfo.UserID, &userInfo.Username, &t.ID, &t.Type, &t.Amount, &t.CreatedAt)
			if err != nil {
				http.Error(w, "Error scanning transaction", http.StatusInternalServerError)
				return
			}

			// Add user info and transaction to the response
			transactions = append(transactions, map[string]interface{}{
				"user_id":  userInfo.UserID,
				"username": userInfo.Username,
				"transaction": map[string]interface{}{
					"id":         t.ID,
					"type":       t.Type,
					"amount":     t.Amount,
					"created_at": t.CreatedAt,
				},
			})
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		if len(transactions) == 0 {
			json.NewEncoder(w).Encode(map[string]string{"message": "No transactions found."})
		} else {
			json.NewEncoder(w).Encode(transactions)
		}
	}
}

// GetUserDetails returns user details along with balance and total number of transactions
func GetUserDetails(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract ID from the path parameters
		vars := mux.Vars(r)
		id := vars["id"]
		if id == "" {
			http.Error(w, "id is required", http.StatusBadRequest)
			return
		}

		// Query to get user details and balance
		var userDetails struct {
			ID       int     `json:"id"`
			Username string  `json:"username"`
			Balance  float64 `json:"balance"`
		}
		err := db.QueryRow(`SELECT id, username, balance FROM users WHERE id = $1`, id).Scan(
			&userDetails.ID, &userDetails.Username, &userDetails.Balance)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "User not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Failed to fetch user details", http.StatusInternalServerError)
			return
		}

		// Query to count total transactions
		var totalTransactions int
		err = db.QueryRow(`SELECT COUNT(*) FROM transactions WHERE user_id = $1`, id).Scan(&totalTransactions)
		if err != nil {
			http.Error(w, "Failed to fetch transaction count", http.StatusInternalServerError)
			return
		}

		// Create the response
		response := map[string]interface{}{
			"id":                 userDetails.ID,
			"username":           userDetails.Username,
			"balance":            userDetails.Balance,
			"total_transactions": totalTransactions,
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// GetTransactionSummary handles the retrieval of the transaction summary for a user
func GetTransactionSummary(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Extract user ID from URL parameters using Gorilla mux
		vars := mux.Vars(r)
		id := vars["id"]
		if id == "" {
			http.Error(w, "id is required", http.StatusBadRequest)
			return
		}

		// Query to get the user details (name and balance)
		var user models.User
		err := db.QueryRow("SELECT id, username, balance FROM users WHERE id = $1", id).Scan(&user.ID, &user.Username, &user.Balance)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "User not found", http.StatusNotFound)
			} else {
				http.Error(w, "Failed to fetch user details", http.StatusInternalServerError)
			}
			return
		}

		// Query to get the transaction count and totals for credit and debit
		var creditCount, debitCount int
		var creditAmount, debitAmount float64

		// Total transactions, credit, and debit amounts
		rows, err := db.Query(`
			SELECT type, amount
			FROM transactions
			WHERE user_id = $1`, id)
		if err != nil {
			http.Error(w, "Failed to fetch transactions", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var tType string
			var amount float64
			err := rows.Scan(&tType, &amount)
			if err != nil {
				http.Error(w, "Error scanning transaction", http.StatusInternalServerError)
				return
			}

			// Count credit/debit transactions and calculate amounts
			if tType == "credit" {
				creditCount++
				creditAmount += amount
			} else if tType == "debit" {
				debitCount++
				debitAmount += amount
			}
		}

		// Prepare the transaction summary
		summary := struct {
			UserID            int     `json:"user_id"`
			Username          string  `json:"username"`
			Balance           float64 `json:"balance"`
			TotalTransactions int     `json:"total_transactions"`
			TotalCredit       struct {
				Count  int     `json:"count"`
				Amount float64 `json:"amount"`
			} `json:"total_credit"`
			TotalDebit struct {
				Count  int     `json:"count"`
				Amount float64 `json:"amount"`
			} `json:"total_debit"`
		}{
			UserID:            user.ID,
			Username:          user.Username,
			Balance:           user.Balance,
			TotalTransactions: creditCount + debitCount,
			TotalCredit: struct {
				Count  int     `json:"count"`
				Amount float64 `json:"amount"`
			}{
				Count:  creditCount,
				Amount: creditAmount,
			},
			TotalDebit: struct {
				Count  int     `json:"count"`
				Amount float64 `json:"amount"`
			}{
				Count:  debitCount,
				Amount: debitAmount,
			},
		}

		// Send the JSON response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(summary)
	}
}
