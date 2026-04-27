package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

// User represents a platform user with RBAC roles.
type User struct {
	ID           string `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"`
	Role         string `json:"role"`
}

// VM represents a managed Virtual Machine entity.
type VM struct {
	ID         string    `json:"id"`
	Name               string    `json:"name"`
	Host               string    `json:"host"`
	ManagementUsername string    `json:"management_username"`
	AuthType   string    `json:"auth_type"`
	Credential string    `json:"-"`        // AES Encrypted
	KeyID      string    `json:"key_id,omitempty"` // optional FK to ssh_keys
	Tags       []string  `json:"tags"`
	Status     string    `json:"status"`
	ProvisioningStatus string `json:"provisioning_status"`
	AuthorizedUsers    []string `json:"authorized_users"`
	CreatedAt  time.Time `json:"created_at"`
}

// AuditLogEntry represents a secure record of an event.
type AuditLogEntry struct {
	ID        int       `json:"id"`
	EventType string    `json:"event_type"`
	Details   string    `json:"details"`
	Timestamp time.Time `json:"timestamp"`
}

// CreateVM persists a new VM record to the database.
func CreateVM(db *sql.DB, vm VM) error {
	tagsJSON, _ := json.Marshal(vm.Tags)
	_, err := db.Exec(
		`INSERT INTO vms (id, name, host, management_username, auth_type, credential, key_id, tags, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		vm.ID, vm.Name, vm.Host, vm.ManagementUsername, vm.AuthType, vm.Credential,
		nullableStr(vm.KeyID), string(tagsJSON), vm.Status)
	return err
}

// UpdateVM modifies an existing VM record by ID, including optional key_id.
func UpdateVM(db *sql.DB, id, name, host, managementUsername, authType, credential, keyID string, tags []string) error {
	tagsJSON, _ := json.Marshal(tags)
	_, err := db.Exec(
		`UPDATE vms SET name=?, host=?, management_username=?, auth_type=?, credential=?, key_id=?, tags=? WHERE id=?`,
		name, host, managementUsername, authType, credential, nullableStr(keyID), string(tagsJSON), id)
	return err
}

// UpdateVMStatus updates only the connectivity status of a VM and records the heartbeat.
func UpdateVMStatus(db *sql.DB, id, status string) error {
	if status == "online" {
		_, err := db.Exec(`UPDATE vms SET status=?, last_seen_at=CURRENT_TIMESTAMP WHERE id=?`, status, id)
		return err
	}
	_, err := db.Exec(`UPDATE vms SET status=? WHERE id=?`, status, id)
	return err
}

// nullableStr converts an empty string to nil so SQLite stores NULL instead of "".
func nullableStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// GetVMs retrieves all registered VMs (excluding credentials) with their provisioning status.
func GetVMs(db *sql.DB) ([]VM, error) {
	rows, err := db.Query(`
		SELECT v.id, v.name, v.host, v.management_username, v.auth_type, v.tags, v.status, v.created_at, 
		       COALESCE(pr.status, 'not_provisioned') as prov_status
		FROM vms v
		LEFT JOIN provisioning_records pr ON v.id = pr.vm_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vms []VM
	for rows.Next() {
		var vm VM
		var tagsStr sql.NullString
		if err := rows.Scan(&vm.ID, &vm.Name, &vm.Host, &vm.ManagementUsername, &vm.AuthType, &tagsStr, &vm.Status, &vm.CreatedAt, &vm.ProvisioningStatus); err != nil {
			return nil, err
		}
		if tagsStr.Valid {
			json.Unmarshal([]byte(tagsStr.String), &vm.Tags)
		} else {
			vm.Tags = []string{}
		}

		// Fetch authorized users for this VM
		users, _ := GetUsersByVM(db, vm.ID)
		vm.AuthorizedUsers = users

		vms = append(vms, vm)
	}
	return vms, nil
}

// GetVMByID retrieves a single VM by its UUID, including the encrypted credential and provisioning status.
func GetVMByID(db *sql.DB, id string) (*VM, error) {
	var vm VM
	var tagsStr sql.NullString
	var keyID sql.NullString
	err := db.QueryRow(`
		SELECT v.id, v.name, v.host, v.management_username, v.auth_type, v.credential, v.key_id, v.tags, v.status, v.created_at,
		       COALESCE(pr.status, 'not_provisioned') as prov_status
		FROM vms v
		LEFT JOIN provisioning_records pr ON v.id = pr.vm_id
		WHERE v.id = ?`, id).
		Scan(&vm.ID, &vm.Name, &vm.Host, &vm.ManagementUsername, &vm.AuthType, &vm.Credential, &keyID, &tagsStr, &vm.Status, &vm.CreatedAt, &vm.ProvisioningStatus)
	if err != nil {
		return nil, err
	}
	if keyID.Valid {
		vm.KeyID = keyID.String
	}
	if tagsStr.Valid {
		json.Unmarshal([]byte(tagsStr.String), &vm.Tags)
	} else {
		vm.Tags = []string{}
	}

	// Fetch authorized users
	users, _ := GetUsersByVM(db, id)
	vm.AuthorizedUsers = users

	return &vm, nil
}

// DeleteVM removes a VM and its associated provisioning record.
func DeleteVM(db *sql.DB, id string) error {
	// 1. Delete provisioning record
	_, _ = db.Exec(`DELETE FROM provisioning_records WHERE vm_id = ?`, id)
	
	// 2. Delete VM record
	_, err := db.Exec(`DELETE FROM vms WHERE id = ?`, id)
	return err
}

// CreateAuditLog inserts a new audit event into the database.
func CreateAuditLog(db *sql.DB, eventType, details string) error {
	_, err := db.Exec(`INSERT INTO audit_log_entries (event_type, details) VALUES (?, ?)`, eventType, details)
	return err
}

// AuditLog represents an entry in the audit log.
type AuditLog struct {
	ID        int    `json:"id"`
	EventType string `json:"event_type"`
	Details   string `json:"details"`
	Timestamp string `json:"timestamp"`
}

// GetAuditLogs retrieves paginated audit logs from the database.
func GetAuditLogs(db *sql.DB, limit, offset int) ([]AuditLog, int, error) {
	var total int
	err := db.QueryRow(`SELECT count(*) FROM audit_log_entries`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := db.Query(`SELECT id, event_type, details, timestamp FROM audit_log_entries ORDER BY timestamp DESC LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var log AuditLog
		if err := rows.Scan(&log.ID, &log.EventType, &log.Details, &log.Timestamp); err != nil {
			return nil, 0, err
		}
		logs = append(logs, log)
	}
	if logs == nil {
		logs = []AuditLog{}
	}
	return logs, total, nil
}

// AgentToken represents an OTT used for agent enrollment.
type AgentToken struct {
	ID        string     `json:"id"`
	TokenHash string     `json:"-"`
	MachineID string     `json:"machine_id"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
}

// AgentCertificate tracks issued mTLS certificates.
type AgentCertificate struct {
	SerialNumber     string     `json:"serial_number"`
	MachineID        string     `json:"machine_id"`
	IssuedAt         time.Time  `json:"issued_at"`
	ExpiresAt        time.Time  `json:"expires_at"`
	RevokedAt        *time.Time `json:"revoked_at,omitempty"`
	RevocationReason string     `json:"revocation_reason,omitempty"`
}

// CreateAgentToken persists a new OTT.
func CreateAgentToken(db *sql.DB, token AgentToken) error {
	_, err := db.Exec(
		`INSERT INTO agent_tokens (id, token_hash, machine_id, created_at, expires_at) VALUES (?, ?, ?, ?, ?)`,
		token.ID, token.TokenHash, token.MachineID, token.CreatedAt, token.ExpiresAt)
	return err
}

// GetAgentTokenByHash retrieves a token by its SHA256 hash.
func GetAgentTokenByHash(db *sql.DB, hash string) (*AgentToken, error) {
	var t AgentToken
	var usedAt *time.Time
	err := db.QueryRow(`SELECT id, token_hash, machine_id, created_at, expires_at, used_at FROM agent_tokens WHERE token_hash = ?`, hash).
		Scan(&t.ID, &t.TokenHash, &t.MachineID, &t.CreatedAt, &t.ExpiresAt, &usedAt)
	if err != nil {
		return nil, err
	}
	t.UsedAt = usedAt
	return &t, nil
}

// MarkTokenUsed records the timestamp when an OTT was consumed.
func MarkTokenUsed(db *sql.DB, id string) error {
	_, err := db.Exec(`UPDATE agent_tokens SET used_at = ? WHERE id = ?`, time.Now(), id)
	return err
}

// CreateAgentCertificate records a newly issued certificate.
func CreateAgentCertificate(db *sql.DB, cert AgentCertificate) error {
	_, err := db.Exec(
		`INSERT INTO agent_certificates (serial_number, machine_id, issued_at, expires_at) VALUES (?, ?, ?, ?)`,
		cert.SerialNumber, cert.MachineID, cert.IssuedAt, cert.ExpiresAt)
	return err
}

// IsCertificateRevoked checks if a certificate serial is in the CRL.
func IsCertificateRevoked(db *sql.DB, serial string) (bool, error) {
	var count int
	err := db.QueryRow(`SELECT count(*) FROM agent_certificates WHERE serial_number = ? AND revoked_at IS NOT NULL`, serial).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// RevokeAgentCertificates marks all certificates for a machine as revoked.
func RevokeAgentCertificates(db *sql.DB, machineID string) error {
	_, err := db.Exec(`UPDATE agent_certificates SET revoked_at = ?, revocation_reason = 'Administrative revocation' WHERE machine_id = ?`, time.Now(), machineID)
	return err
}

