package models

import (
	"database/sql"
	"time"
)

// SSHKey represents an encrypted SSH key pair stored in the vault.
type SSHKey struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	PublicKey  string    `json:"public_key"`
	// PrivateKey is NEVER returned in API responses; only used internally
	PrivateKey string    `json:"-"`
	CreatedAt  time.Time `json:"created_at"`
}

// SSHKeySummary is the safe API response shape (no private key).
type SSHKeySummary struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	PublicKey string    `json:"public_key"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateSSHKey persists a new encrypted key pair to the vault.
func CreateSSHKey(db *sql.DB, key SSHKey) error {
	_, err := db.Exec(
		`INSERT INTO ssh_keys (id, name, private_key, public_key, created_at) VALUES (?, ?, ?, ?, ?)`,
		key.ID, key.Name, key.PrivateKey, key.PublicKey, key.CreatedAt,
	)
	return err
}

// GetSSHKeys lists all key summaries (without private key data).
func GetSSHKeys(db *sql.DB) ([]SSHKeySummary, error) {
	rows, err := db.Query(`SELECT id, name, public_key, created_at FROM ssh_keys ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []SSHKeySummary
	for rows.Next() {
		var k SSHKeySummary
		if err := rows.Scan(&k.ID, &k.Name, &k.PublicKey, &k.CreatedAt); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, nil
}

// GetSSHKeyByID retrieves the full key record including the encrypted private key.
// Only used internally (SSH client decryption) — never exposed via API.
func GetSSHKeyByID(db *sql.DB, id string) (*SSHKey, error) {
	var k SSHKey
	err := db.QueryRow(
		`SELECT id, name, private_key, public_key, created_at FROM ssh_keys WHERE id = ?`, id,
	).Scan(&k.ID, &k.Name, &k.PrivateKey, &k.PublicKey, &k.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &k, nil
}

// DeleteSSHKey removes a key from the vault by ID.
func DeleteSSHKey(db *sql.DB, id string) error {
	_, err := db.Exec(`DELETE FROM ssh_keys WHERE id = ?`, id)
	return err
}
