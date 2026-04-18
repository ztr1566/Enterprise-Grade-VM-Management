package integration

import (
	"database/sql"
	"errors"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"backend/internal/ssh"
	"backend/internal/models"
	"backend/internal/crypto"
	"backend/internal/audit"
	"io"
	gossh "golang.org/x/crypto/ssh"
)

// MockSession implements ssh.SSHSession
type MockSession struct {
	RunCmdFunc func(cmd string) ([]byte, error)
	Closed     bool
}


func (m *MockSession) RunCmd(cmd string) ([]byte, error) {
	return m.RunCmdFunc(cmd)
}

func (m *MockSession) Run(cmd string) error {
	_, err := m.RunCmdFunc(cmd)
	return err
}

func (m *MockSession) RequestPty(term string, h, w int, modes gossh.TerminalModes) error { return nil }
func (m *MockSession) StdinPipe() (io.WriteCloser, error)                   { return nil, nil }
func (m *MockSession) StdoutPipe() (io.Reader, error)                       { return nil, nil }
func (m *MockSession) Shell() error                                         { return nil }
func (m *MockSession) Start(cmd string) error                               { return nil }
func (m *MockSession) Wait() error                                          { return nil }
func (m *MockSession) WindowChange(rows, cols int) error                    { return nil }

func (m *MockSession) Close() {
	m.Closed = true
}


func setupTestDB(t *testing.T) *sql.DB {
	// Initialize logger for tests
	audit.InitLogger()

	// Ensure AES_KEY is set for tests
	if os.Getenv("AES_KEY") == "" {
		os.Setenv("AES_KEY", "YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXoxMjM0NTY=") // 32 bytes base64
	}

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	// Apply migrations
	migrations := []string{
		"../../internal/db/migrations/001_init.sql",
		"../../internal/db/migrations/002_ssh_keys.sql",
		"../../internal/db/migrations/003_provisioning.sql",
		"../../internal/db/migrations/004_vm_access.sql",
		"../../internal/db/migrations/005_identity_refactor.sql",
	}

	for _, m := range migrations {
		content, err := os.ReadFile(m)
		if err != nil {
			t.Fatalf("failed to read migration %s: %v", m, err)
		}
		if _, err := db.Exec(string(content)); err != nil {
			t.Fatalf("failed to apply migration %s: %v", m, err)
		}
	}

	return db
}

func TestProvisionVM_Success(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// 1. Setup VM and provisioning record
	vmID := "test-vm-1"
	aesKey := crypto.MustGetAESKey()
	encCred, _ := crypto.Encrypt([]byte("password"), aesKey)
	
	_, err := db.Exec(`INSERT INTO vms (id, name, host, management_username, auth_type, credential, status) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		vmID, "Test VM", "1.2.3.4", "admin", "password", encCred, "online")
	if err != nil {
		t.Fatalf("failed to insert vm: %v", err)
	}
	
	models.CreateProvisioningRecord(db, vmID)

	// 2. Mock SSH session
	oldNewSession := ssh.NewSession
	defer func() { ssh.NewSession = oldNewSession }()

	ssh.NewSession = func(db *sql.DB, id string, loginAs ...string) (ssh.SSHSession, error) {
		return &MockSession{
			RunCmdFunc: func(cmd string) ([]byte, error) {
				return []byte("success"), nil
			},
		}, nil
	}

	// 3. Execute ProvisionVM
	err = ssh.ProvisionVM(db, vmID)
	if err != nil {
		t.Fatalf("ProvisionVM failed: %v", err)
	}

	// 4. Verify DB state
	record, err := models.GetProvisioningRecord(db, vmID)
	if err != nil {
		t.Fatalf("failed to get record: %v", err)
	}
	if record.Status != "provisioned" {
		t.Errorf("expected status 'provisioned', got '%s'", record.Status)
	}

	// 5. Verify audit logs
	var count int
	db.QueryRow(`SELECT COUNT(*) FROM audit_log_entries WHERE event_type = 'VM_PROVISIONED'`).Scan(&count)
	if count != 1 {
		t.Errorf("expected 1 VM_PROVISIONED audit log, got %d", count)
	}
}

func TestProvisionVM_VisudoFailure(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// 1. Setup VM
	vmID := "test-vm-2"
	aesKey := crypto.MustGetAESKey()
	encCred, _ := crypto.Encrypt([]byte("password"), aesKey)
	db.Exec(`INSERT INTO vms (id, name, host, management_username, auth_type, credential, status) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		vmID, "Test VM", "1.2.3.4", "admin", "password", encCred, "online")
	models.CreateProvisioningRecord(db, vmID)

	// 2. Mock SSH session failure (simulating visudo error)
	oldNewSession := ssh.NewSession
	defer func() { ssh.NewSession = oldNewSession }()

	ssh.NewSession = func(db *sql.DB, id string, loginAs ...string) (ssh.SSHSession, error) {
		return &MockSession{
			RunCmdFunc: func(cmd string) ([]byte, error) {
				if cmd == "sudo bash /tmp/setup_vm_sudoers.sh admin" {
					return []byte("visudo: /etc/sudoers.d/vm-platform-admin: parse error"), errors.New("exit status 1")
				}
				return []byte("ok"), nil
			},
		}, nil
	}

	// 3. Execute ProvisionVM
	err := ssh.ProvisionVM(db, vmID)
	if err == nil {
		t.Fatal("expected ProvisionVM to fail")
	}

	// 4. Verify DB state
	record, err := models.GetProvisioningRecord(db, vmID)
	if err != nil {
		t.Fatalf("failed to get record: %v", err)
	}
	if record.Status != "failed" {
		t.Errorf("expected status 'failed', got '%s'", record.Status)
	}
	if record.ErrorMessage == nil || *record.ErrorMessage == "" {
		t.Error("expected error message in record")
	}

	// 5. Verify audit logs
	var count2 int
	db.QueryRow(`SELECT COUNT(*) FROM audit_log_entries WHERE event_type = 'VM_USER_PROVISIONING_FAILED'`).Scan(&count2)
	if count2 != 1 {
		t.Errorf("expected 1 VM_USER_PROVISIONING_FAILED audit log, got %d", count2)
	}
}

