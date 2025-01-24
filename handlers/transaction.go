package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// Perform Debit and Credit Transaction locking using transactions
func TransactionStart(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Read the user ID and username from the body of the request
		var user struct {
			ID       int    `json:"id"`
			Username string `json:"username"`
		}
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		// Check if the user exists in the database
		var userExists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND username = $2)", user.ID, user.Username).Scan(&userExists)
		if err != nil || !userExists {
			http.Error(w, "User does not exist", http.StatusNotFound)
			return
		}

		// Start a database transaction
		tx, err := db.Begin()
		if err != nil {
			http.Error(w, "Failed to begin transaction", http.StatusInternalServerError)
			return
		}

		// Lock the shared wallet row using SELECT ... FOR UPDATE
		var balance float64
		err = tx.QueryRow("SELECT balance FROM shared_wallet WHERE id = 1 FOR UPDATE").Scan(&balance)
		if err != nil {
			tx.Rollback() // Rollback transaction on error
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}



		// Ask for transaction details in termial
		fmt.Println("\n\nHi", user.Username)
		fmt.Println("Enter operation: 'debit' or 'credit'")
		var operation string
		fmt.Scanln(&operation)

		operation = strings.ToLower(operation)
		if operation != "debit" && operation != "credit" {
			http.Error(w, "Invalid operation", http.StatusBadRequest)
			fmt.Println("Invalid operation")
			return
		}

		fmt.Println("Enter amount:")
		var amount float64
		_, err = fmt.Scanln(&amount)
		if err != nil || amount <= 0 {
			http.Error(w, "Invalid amount", http.StatusBadRequest)
			return
		}



		// Perform the transaction logic: Debit or Credit
		if operation == "debit" {
			if amount > balance {
				http.Error(w, "Insufficient balance", http.StatusBadRequest)
				fmt.Println("Insufficient balance")
				return
			}
			// Debit operation
			_, err = tx.Exec("UPDATE shared_wallet SET balance = balance - $1 WHERE id = 1", amount)
		} else {
			// Credit operation
			_, err = tx.Exec("UPDATE shared_wallet SET balance = balance + $1 WHERE id = 1", amount)
		}
		if err != nil {
			tx.Rollback() // Rollback the transaction if error occurs
			http.Error(w, "Error performing the transaction", http.StatusInternalServerError)
			return
		}



		// Log the transaction (Insert into the transactions table)
		_, err = tx.Exec(
			"INSERT INTO transactions (user_id, user_name, type, amount) VALUES ($1, $2, $3, $4)",
			user.ID, user.Username, operation, amount,
		)
		if err != nil {
			tx.Rollback() // Rollback transaction on error
			http.Error(w, "Error logging transaction", http.StatusInternalServerError)
			return
		}

		// Commit the transaction
		if err := tx.Commit(); err != nil {
			http.Error(w, "Error committing transaction", http.StatusInternalServerError)
			return
		}

		// Return success response
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Transaction completed successfully\nOperation: %s\nAmount: %f\n", operation, amount)
		fmt.Printf("Transaction completed successfully\nOperation: %s\nAmount: %f\n", operation, amount)
	}
}

// GetTransactions fetches all transactions for a given user by their ID
func GetTransactions(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Extract ID from the path parameters
		vars := mux.Vars(r)
		userID := vars["id"]
		if userID == "" {
			http.Error(w, "id is required", http.StatusBadRequest)
			return
		}

		// Prepare a query to get all transactions for the user by their ID
		rows, err := db.Query("SELECT transaction_id, user_id, user_name, type, amount, created_at FROM transactions WHERE user_id = $1", userID)
		if err != nil {
			http.Error(w, "Error fetching transactions", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// Create a slice to hold the transactions
		var transactions []struct {
			TransactionID int     `json:"transaction_id"`
			UserID        int     `json:"user_id"`
			UserName      string  `json:"user_name"`
			Type          string  `json:"type"`
			Amount        float64 `json:"amount"`
			CreatedAt     string  `json:"created_at"`
		}

		// Iterate over the rows and add each transaction to the slice
		for rows.Next() {
			var t struct {
				TransactionID int     `json:"transaction_id"`
				UserID        int     `json:"user_id"`
				UserName      string  `json:"user_name"`
				Type          string  `json:"type"`
				Amount        float64 `json:"amount"`
				CreatedAt     string  `json:"created_at"`
			}
			if err := rows.Scan(&t.TransactionID, &t.UserID, &t.UserName, &t.Type, &t.Amount, &t.CreatedAt); err != nil {
				http.Error(w, "Error reading transaction data", http.StatusInternalServerError)
				return
			}
			transactions = append(transactions, t)
		}

		// Handle case where no transactions are found
		if len(transactions) == 0 {
			http.Error(w, "No transactions found for the user", http.StatusNotFound)
			return
		}

		// Convert the transactions slice into JSON format and send the response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(transactions); err != nil {
			http.Error(w, "Error encoding transactions to JSON", http.StatusInternalServerError)
		}
	}
}

// GetTransactionSummary:- Get the transaction summary for a particular user
func GetTransactionSummary(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user ID from URL
		id := mux.Vars(r)["id"]
		if id == "" {
			http.Error(w, "id is required", http.StatusBadRequest)
			return
		}
		// Combine single query to get the necessary transaction data.
		var (
			totalTransactions   int
			totalCredits        int
			totalDebits         int
			totalCreditedAmount float64
			totalDebitedAmount  float64
		)
		query := `
            SELECT 
                (SELECT COUNT(*) FROM transactions WHERE user_id = $1),
                (SELECT COUNT(*) FROM transactions WHERE user_id = $1 AND type = 'credit'),
                (SELECT COUNT(*) FROM transactions WHERE user_id = $1 AND type = 'debit'),
                (SELECT COALESCE(SUM(amount), 0) FROM transactions WHERE user_id = $1 AND type = 'credit'),
                (SELECT COALESCE(SUM(amount), 0) FROM transactions WHERE user_id = $1 AND type = 'debit')
        `
		err := db.QueryRow(query, id).Scan(
			&totalTransactions,
			&totalCredits,
			&totalDebits,
			&totalCreditedAmount,
			&totalDebitedAmount,
		)
		if err != nil {
			log.Printf("Error fetching transaction summary: %v", err)
			http.Error(w, "Error fetching transaction summary", http.StatusInternalServerError)
			return
		}

		// Create a struct to hold the summary information.
		summary := struct {
			TotalTransactions   int     `json:"total_transactions"`
			TotalCredits        int     `json:"total_credits"`
			TotalCreditedAmount float64 `json:"total_credited_amount"`
			TotalDebits         int     `json:"total_debits"`
			TotalDebitedAmount  float64 `json:"total_debited_amount"`
		}{
			TotalTransactions:   totalTransactions,
			TotalCredits:        totalCredits,
			TotalCreditedAmount: totalCreditedAmount,
			TotalDebits:         totalDebits,
			TotalDebitedAmount:  totalDebitedAmount,
		}

		// Set the response content type to JSON.
		w.Header().Set("Content-Type", "application/json")

		// Encode the summary struct to JSON and send as response.
		if err := json.NewEncoder(w).Encode(summary); err != nil {
			log.Printf("Error encoding summary: %v", err)
			http.Error(w, "Error generating response", http.StatusInternalServerError)
		}
	}
}

// GetWalletDetails:- Get The wallet details ie. Wallet Balance and Wallet Lock Status
func GetWalletDetails(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Variables to hold wallet details
		var balance float64
		var createdAt time.Time

		// Query to fetch wallet details
		err := db.QueryRow("SELECT balance, created_at FROM shared_wallet WHERE id = 1").Scan(&balance, &createdAt)
		if err != nil {
			log.Printf("Error fetching wallet details: %v", err)
			http.Error(w, "Error fetching wallet details", http.StatusInternalServerError)
			return
		}

		// Struct to hold the response data
		walletDetails := struct {
			Balance   float64   `json:"balance"`
			CreatedAt time.Time `json:"created_at"`
		}{
			Balance:   balance,
			CreatedAt: createdAt,
		}

		// Set response content type
		w.Header().Set("Content-Type", "application/json")

		// Encode the wallet details into JSON and return
		if err := json.NewEncoder(w).Encode(walletDetails); err != nil {
			log.Printf("Error encoding wallet details: %v", err)
			http.Error(w, "Error generating response", http.StatusInternalServerError)
		}
	}
}
