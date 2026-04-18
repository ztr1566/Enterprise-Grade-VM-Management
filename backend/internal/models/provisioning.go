package models

import (
	"database/sql"
	"time"
)

// ProvisioningRecord tracks the provisioning state of a target VM.
type ProvisioningRecord struct {
	VMID            string     `json:"vm_id"`
	Status          string     `json:"status"`
	LastAttemptAt   *time.Time `json:"last_attempt_at"`
	ErrorMessage    *string    `json:"error_message"`
	ProvisionedUser *string    `json:"provisioned_user"`
}

// CreateProvisioningRecord initializes a new record with 'not_provisioned' status.
func CreateProvisioningRecord(db *sql.DB, vmID string) error {
	_, err := db.Exec(
		`INSERT INTO provisioning_records (vm_id, status) VALUES (?, 'not_provisioned')`,
		vmID)
	return err
}

// GetProvisioningRecord retrieves the provisioning state for a specific VM.
func GetProvisioningRecord(db *sql.DB, vmID string) (*ProvisioningRecord, error) {
	var pr ProvisioningRecord
	var lastAttempt sql.NullTime
	var errMsg, provUser sql.NullString

	err := db.QueryRow(
		`SELECT vm_id, status, last_attempt_at, error_message, provisioned_user FROM provisioning_records WHERE vm_id = ?`,
		vmID).Scan(&pr.VMID, &pr.Status, &lastAttempt, &errMsg, &provUser)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if lastAttempt.Valid {
		pr.LastAttemptAt = &lastAttempt.Time
	}
	if errMsg.Valid {
		pr.ErrorMessage = &errMsg.String
	}
	if provUser.Valid {
		pr.ProvisionedUser = &provUser.String
	}

	return &pr, nil
}

// UpdateProvisioningStatus updates the status and related details of a provisioning attempt.
func UpdateProvisioningStatus(db *sql.DB, vmID, status string, errDetail *string, user *string) error {
	now := time.Now()
	_, err := db.Exec(
		`UPDATE provisioning_records SET status = ?, last_attempt_at = ?, error_message = ?, provisioned_user = ? WHERE vm_id = ?`,
		status, now, nullableStrPtr(errDetail), nullableStrPtr(user), vmID)
	return err
}

// nullableStrPtr is a helper for optional string pointers to database/sql interfaces.
func nullableStrPtr(s *string) interface{} {
	if s == nil {
		return nil
	}
	return *s
}
