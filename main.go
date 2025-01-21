package main

import (
	"fmt"
	"log"
	"net/http"
	"wallet-system/handlers"
	"wallet-system/middleware" // Import the middleware package
	database "wallet-system/storage"

	"github.com/gorilla/mux"
)

func main() {
	// Connect to the database
	db := database.Connect()
	defer db.Close()

	r := mux.NewRouter()

	// User-related endpoints
	r.HandleFunc("/user/create", handlers.CreateUser(db)).Methods("POST")          // Create User
	r.HandleFunc("/user/{id}/update", handlers.UpdateUser(db)).Methods("PUT")      // Update User by ID
	r.HandleFunc("/user/{id}/details", handlers.GetUserDetails(db)).Methods("GET") // Get User details by ID
	r.HandleFunc("/user/{id}/delete", handlers.DeleteUser(db)).Methods("DELETE")   // Get User details by ID

	// Transaction-related endpoints
	r.HandleFunc("/transaction", handlers.TransactionStart(db)).Methods("POST")                       // Transaction (Debit/Credit)
	r.HandleFunc("/transactions/user/{id}", handlers.GetTransactions(db)).Methods("GET")              // Get all transactions for user by ID
	r.HandleFunc("/user/transaction-summary/{id}", handlers.GetTransactionSummary(db)).Methods("GET") // Get transaction summary for user by ID
	r.HandleFunc("/wallet", handlers.GetWalletDetails(db)).Methods("GET")                             // Get Wallet Details

	// Health check endpoint
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Health check passed")
	}).Methods("GET")

	// Wrap router with logging middleware
	loggedRouter := middleware.LoggingMiddleware(r)

	// Start the server
	log.Println("Server is running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", loggedRouter))
}
