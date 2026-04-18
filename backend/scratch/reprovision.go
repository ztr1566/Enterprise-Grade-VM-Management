package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"

	"backend/internal/ssh"
)

func main() {
	_ = godotenv.Load()
	db, err := sql.Open("sqlite3", "vm-manager.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT id FROM vms")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			continue
		}
		fmt.Println("Provisioning VM:", id)
		err := ssh.ProvisionVM(db, id)
		if err != nil {
			fmt.Printf("Failed to provision %s: %v\n", id, err)
		} else {
			fmt.Printf("Successfully provisioned %s\n", id)
		}
	}
}
