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
	r.HandleFunc("/user/{id}/delete", handlers.DeleteUser(db)).Methods("DELETE")   // Delete User by ID
	r.HandleFunc("/user/details/{id}", handlers.GetUserDetails(db)).Methods("GET") // Get User details by ID

	// Transaction-related endpoints
	r.HandleFunc("/user/credit", handlers.CreditBalance(db)).Methods("POST")                          // Credit balance for user by ID
	r.HandleFunc("/user/debit", handlers.DebitBalance(db)).Methods("POST")                            // Debit balance for user by ID
	r.HandleFunc("/user/transactions/{id}", handlers.GetTransactions(db)).Methods("GET")              // Get all transactions for user by ID
	r.HandleFunc("/user/transaction-summary/{id}", handlers.GetTransactionSummary(db)).Methods("GET") // Get transaction summary for user by ID

	// Health check endpoint
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Health check passed")
	}).Methods("GET")

	// Test endpoint (for example use)
	r.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, Wallet System!")
	}).Methods("GET")

	// Wrap router with logging middleware
	loggedRouter := middleware.LoggingMiddleware(r)

	// Start the server
	log.Println("Server is running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", loggedRouter))
}
