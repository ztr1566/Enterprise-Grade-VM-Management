package models

import (
	"database/sql"
	"fmt"
)

// AddUserToVM grants a specific OS username access to a VM.
func AddUserToVM(db *sql.DB, vmID, username string) error {
	_, err := db.Exec(
		`INSERT INTO vm_access (vm_id, username) VALUES (?, ?) ON CONFLICT(vm_id, username) DO NOTHING`,
		vmID, username)
	if err != nil {
		return fmt.Errorf("adding user to vm access: %w", err)
	}
	return nil
}

// GetUsersByVM retrieves the list of OS usernames permitted for a specific VM.
func GetUsersByVM(db *sql.DB, vmID string) ([]string, error) {
	rows, err := db.Query(`SELECT username FROM vm_access WHERE vm_id = ?`, vmID)
	if err != nil {
		return nil, fmt.Errorf("fetching vm users: %w", err)
	}
	defer rows.Close()

	var users []string
	for rows.Next() {
		var user string
		if err := rows.Scan(&user); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

// RemoveUserFromVM revokes a specific OS username's access from a VM.
func RemoveUserFromVM(db *sql.DB, vmID, username string) error {
	_, err := db.Exec(`DELETE FROM vm_access WHERE vm_id = ? AND username = ?`, vmID, username)
	if err != nil {
		return fmt.Errorf("removing user from vm access: %w", err)
	}
	return nil
}
