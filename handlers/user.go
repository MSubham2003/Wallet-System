package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

// validateName validates a name field (Fname or Lname) based on defined rules.
func validateName(name string, fieldName string) error {
	if name == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	if strings.TrimSpace(name) != name {
		return fmt.Errorf("%s cannot have leading or trailing spaces", fieldName)
	}
	if regexp.MustCompile(`[0-9]`).MatchString(name) {
		return fmt.Errorf("%s cannot contain numbers", fieldName)
	}
	if regexp.MustCompile(`[^a-zA-Z\s]`).MatchString(name) {
		return fmt.Errorf("%s can only contain alphabetic characters and spaces", fieldName)
	}
	if strings.Contains(name, "  ") {
		return fmt.Errorf("%s cannot have consecutive spaces", fieldName)
	}
	if len(name) < 3 || len(name) > 50 {
		return fmt.Errorf("%s must be between 3 and 50 characters long", fieldName)
	}
	return nil
}

// CreateUser: Creates a new user in the database
func CreateUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the JSON body
		var requestData struct {
			Username string `json:"username"`
			Fname    string `json:"fname"`
			Lname    string `json:"lname"`
			Email    string `json:"email"`
		}

		// Decode JSON
		err := json.NewDecoder(r.Body).Decode(&requestData)
		if err != nil {
			http.Error(w, "Invalid JSON input", http.StatusBadRequest)
			return
		}

		// Validate input fields
		if err := validateName(requestData.Fname, "First Name"); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := validateName(requestData.Lname, "Last Name"); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if requestData.Username == "" {
			http.Error(w, "Username is required", http.StatusBadRequest)
			return
		}

		// Validate input: Email
		if requestData.Email == "" || !regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`).MatchString(requestData.Email) {
			http.Error(w, "Valid email is required", http.StatusBadRequest)
			return
		}

		// Insert user into the database
		query := `INSERT INTO users (username, fname, lname, email) VALUES ($1, $2, $3, $4) RETURNING id`
		var id int
		err = db.QueryRow(query, requestData.Username, requestData.Fname, requestData.Lname, requestData.Email).Scan(&id)
		if err != nil {
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			fmt.Println("Error inserting user:", err)
			return
		}

		// Respond with success
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "User created successfully with ID: %d\n", id)
	}
}

// UpdateUser:- updates a user's details (e.g., username, fname, lname, email)
func UpdateUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user ID from URL
		id := mux.Vars(r)["id"]
		if id == "" {
			http.Error(w, "ID is required", http.StatusBadRequest)
			return
		}

		// Parse the JSON body
		var requestData struct {
			Username string `json:"username"`
			Fname    string `json:"fname"`
			Lname    string `json:"lname"`
			Email    string `json:"email"`
		}
		if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
			http.Error(w, "Invalid JSON input", http.StatusBadRequest)
			return
		}

		// Validate inputs
		if requestData.Username == "" || requestData.Fname == "" || requestData.Lname == "" {
			http.Error(w, "Username, Fname, and Lname are required", http.StatusBadRequest)
			return
		}

		// Validate first name
		if err := validateName(requestData.Fname, "First name"); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Validate last name
		if err := validateName(requestData.Lname, "Last name"); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Validate input: Email
		if requestData.Email == "" || !regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`).MatchString(requestData.Email) {
			http.Error(w, "Valid email is required", http.StatusBadRequest)
			return
		}

		// Update the user in the database
		query := `UPDATE users SET username = $1, fname = $2, lname = $3, email = $4 WHERE id = $5`
		_, err := db.Exec(query, requestData.Username, requestData.Fname, requestData.Lname, requestData.Email, id)
		if err != nil {
			http.Error(w, "Failed to update user", http.StatusInternalServerError)
			return
		}

		// Send a success response
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "User with ID %s updated successfully", id)
	}
}

// Get User Details
func GetUserDetails(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract ID from the path parameters
		vars := mux.Vars(r)
		id := vars["id"]
		if id == "" {
			http.Error(w, "id is required", http.StatusBadRequest)
			return
		}
		userID, err := strconv.Atoi(id)
		if err != nil {
			http.Error(w, "Invalid User ID", http.StatusBadRequest)
			return
		}

		// Fetch user details from the database
		var user struct {
			ID                int    `json:"id"`
			Username          string `json:"username"`
			FirstName         string `json:"fname"`
			LastName          string `json:"lname"`
			Email             string `json:"email"`
			TotalTransactions int    `json:"total_transactions"`
		}

		query := `
			SELECT 
				u.id, u.username, u.fname, u.lname, u.email, 
				(SELECT COUNT(*) FROM transactions t WHERE t.user_id = u.id) AS total_transactions
			FROM users u
			WHERE u.id = $1
		`

		err = db.QueryRow(query, userID).Scan(
			&user.ID,
			&user.Username,
			&user.FirstName,
			&user.LastName,
			&user.Email,
			&user.TotalTransactions,
		)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "User not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Error fetching user details", http.StatusInternalServerError)
			return
		}

		// Return the user details as JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	}
}

// DeleteUser:- Delete user by ID
func DeleteUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user ID from URL parameter
		vars := mux.Vars(r)
		userID := vars["id"]

		// Check if the user exists
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", userID).Scan(&exists)
		if err != nil {
			log.Printf("Error checking user existence: %v", err)
			http.Error(w, "Error checking user existence", http.StatusInternalServerError)
			return
		}

		// If the user doesn't exist, return a "user not found" message
		if !exists {
			http.Error(w, fmt.Sprintf("User not found with ID %s", userID), http.StatusNotFound)
			return
		}

		// Delete the user from the users table
		_, err = db.Exec("DELETE FROM users WHERE id = $1", userID)
		if err != nil {
			log.Printf("Error deleting user: %v", err)
			http.Error(w, "Error deleting user", http.StatusInternalServerError)
			return
		}

		// Optionally, log the deletion or return a response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("User deleted, transactions kept"))
	}
}
