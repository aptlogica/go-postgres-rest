package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aptlogica/go-postgres-rest/pkg/config"
	"github.com/aptlogica/go-postgres-rest/pkg/database/interfaces"
	"github.com/aptlogica/go-postgres-rest/pkg/database/postgres"
)

// Constants for HTTP headers and error messages
const (
	contentTypeHeader = "Content-Type"
	contentTypeJSON   = "application/json"
	dbErrorMessage    = "Database error: %v"
	userNotFoundMsg   = "User not found"
)

// User represents our sample data model
type User struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

var db interfaces.DB

func main() {
	fmt.Println("=== Go PostgreSQL REST - Basic CRUD Example ===")

	// Database configuration
	cfg := config.DatabaseConfig{
		Host:            getEnv("DB_HOST", "localhost"),
		Port:            config.ParseInt(getEnv("DB_PORT", "5432"), 5432),
		Username:        getEnv("DB_USER", "postgres"),
		Password:        getEnv("DB_PASSWORD", "password"),
		DatabaseName:    getEnv("DB_NAME", "sereni_examples"),
		SSLMode:         "disable",
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
	}

	// Initialize database connection
	var err error
	db, err = postgres.Connect(&cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	fmt.Printf("✅ Connected to PostgreSQL: %s@%s:%d/%s\n",
		cfg.Username, cfg.Host, cfg.Port, cfg.DatabaseName)

	// Create tables if they don't exist
	if err := createTables(); err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}

	// Setup HTTP routes
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/api/users", usersHandler)
	http.HandleFunc("/api/users/", userHandler)

	port := getEnv("API_PORT", "8080")
	fmt.Printf("\n🚀 Server starting on port %s\n", port)
	fmt.Println("\nAvailable endpoints:")
	fmt.Println("  GET    /health           - Health check")
	fmt.Println("  GET    /api/users        - Get all users")
	fmt.Println("  POST   /api/users        - Create new user")
	fmt.Println("  GET    /api/users/{id}   - Get user by ID")
	fmt.Println("  PUT    /api/users/{id}   - Update user by ID")
	fmt.Println("  DELETE /api/users/{id}   - Delete user by ID")

	fmt.Println("\nExample requests:")
	fmt.Printf(`  curl http://localhost:%s/health`+"\n", port)
	fmt.Printf(`  curl -X POST -H "Content-Type: application/json" -d '{"name":"John Doe","email":"john@example.com"}' http://localhost:%s/api/users`+"\n", port)
	fmt.Printf(`  curl http://localhost:%s/api/users`+"\n", port)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func createTables() error {
	query := `
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		
		-- Create trigger to update updated_at automatically
		CREATE OR REPLACE FUNCTION update_updated_at_column()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = CURRENT_TIMESTAMP;
			RETURN NEW;
		END;
		$$ language 'plpgsql';
		
		DROP TRIGGER IF EXISTS update_users_updated_at ON users;
		CREATE TRIGGER update_users_updated_at
			BEFORE UPDATE ON users
			FOR EACH ROW
			EXECUTE FUNCTION update_updated_at_column();
	`

	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create tables: %v", err)
	}

	fmt.Println("✅ Database tables created/verified")
	return nil
}

// Health check endpoint
func healthHandler(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"database":  "connected",
		"version":   "1.0.0",
	}

	w.Header().Set(contentTypeHeader, contentTypeJSON)
	json.NewEncoder(w).Encode(health)
}

// Handle /api/users (GET, POST)
func usersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getAllUsers(w, r)
	case http.MethodPost:
		createUser(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Handle /api/users/{id} (GET, PUT, DELETE)
func userHandler(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/users/")
	if path == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		getUserByID(w, r, id)
	case http.MethodPut:
		updateUser(w, r, id)
	case http.MethodDelete:
		deleteUser(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Get all users
func getAllUsers(w http.ResponseWriter, r *http.Request) {
	query := `SELECT id, name, email, created_at, updated_at FROM users ORDER BY created_at DESC`

	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, fmt.Sprintf(dbErrorMessage, err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			http.Error(w, fmt.Sprintf("Scan error: %v", err), http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}

	if users == nil {
		users = []User{} // Return empty array instead of null
	}

	w.Header().Set(contentTypeHeader, contentTypeJSON)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":  users,
		"count": len(users),
	})
}

// Create new user
func createUser(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if user.Name == "" || user.Email == "" {
		http.Error(w, "Name and email are required", http.StatusBadRequest)
		return
	}

	query := `
		INSERT INTO users (name, email) 
		VALUES ($1, $2) 
		RETURNING id, name, email, created_at, updated_at
	`

	err := db.QueryRow(query, user.Name, user.Email).Scan(
		&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			http.Error(w, "Email already exists", http.StatusConflict)
		} else {
			http.Error(w, fmt.Sprintf(dbErrorMessage, err), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set(contentTypeHeader, contentTypeJSON)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// Get user by ID
func getUserByID(w http.ResponseWriter, r *http.Request, id int) {
	var user User
	query := `SELECT id, name, email, created_at, updated_at FROM users WHERE id = $1`

	err := db.QueryRow(query, id).Scan(
		&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			http.Error(w, userNotFoundMsg, http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf(dbErrorMessage, err), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set(contentTypeHeader, contentTypeJSON)
	json.NewEncoder(w).Encode(user)
}

// Update user by ID
func updateUser(w http.ResponseWriter, r *http.Request, id int) {
	var updateData User
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Check if user exists
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", id).Scan(&exists)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, userNotFoundMsg, http.StatusNotFound)
		return
	}

	query := `
		UPDATE users 
		SET name = $2, email = $3, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $1 
		RETURNING id, name, email, created_at, updated_at
	`

	var user User
	err = db.QueryRow(query, id, updateData.Name, updateData.Email).Scan(
		&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			http.Error(w, "Email already exists", http.StatusConflict)
		} else {
			http.Error(w, fmt.Sprintf(dbErrorMessage, err), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set(contentTypeHeader, contentTypeJSON)
	json.NewEncoder(w).Encode(user)
}

// Delete user by ID
func deleteUser(w http.ResponseWriter, r *http.Request, id int) {
	result, err := db.Exec("DELETE FROM users WHERE id = $1", id)
	if err != nil {
		http.Error(w, fmt.Sprintf(dbErrorMessage, err), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, userNotFoundMsg, http.StatusNotFound)
		return
	}

	w.Header().Set(contentTypeHeader, contentTypeJSON)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User deleted successfully",
		"id":      id,
	})
}

// Helper function to get environment variables with default values
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
