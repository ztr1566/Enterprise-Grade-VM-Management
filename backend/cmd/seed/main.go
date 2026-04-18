package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Connect to "vm-manager.db" in the current directory
	db, err := sql.Open("sqlite3", "vm-manager.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// 1. Delete any existing user with the username 'admin'
	_, err = db.Exec("DELETE FROM users WHERE username = ?", "admin")
	if err != nil {
		log.Fatalf("Failed to delete existing admin user: %v", err)
	}

	// 2. Prepare new admin user data
	username := "admin"
	password := "Admin@123"
	role := "admin"
	id := uuid.New().String()

	// 3. Hash the password using bcrypt
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// 4. Insert the new admin user
	_, err = db.Exec(`
		INSERT INTO users (id, username, password_hash, role) 
		VALUES (?, ?, ?, ?)
	`, id, username, string(hash), role)

	if err != nil {
		log.Fatalf("Failed to insert admin user: %v", err)
	}

	// 5. Print confirmation message
	fmt.Println("User admin created successfully in vm-manager.db")
}
