package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"backend/internal/audit"
	"backend/internal/db"
	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	audit.InitLogger()
	testDB, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to init test db: %v", err)
	}

	// Run migrations in order
	migrations := []string{"001_init.sql", "002_ssh_keys.sql", "003_provisioning.sql", "004_vm_access.sql", "005_identity_refactor.sql"}
	for _, m := range migrations {
		migration, err := os.ReadFile("../../db/migrations/" + m)
		if err != nil {
			t.Fatalf("failed to read migration %s: %v", m, err)
		}
		if _, err := testDB.Exec(string(migration)); err != nil {
			t.Fatalf("failed to run migration %s: %v", m, err)
		}
	}

	return testDB
}

func TestCreateVM(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	handler := &VMHandler{DB: testDB}
	os.Setenv("AES_KEY", "this-is-a-32-byte-key-!!-1234567")

	reqBody, _ := json.Marshal(CreateVMRequest{
		Name:       "Test VM",
		Host:       "1.2.3.4",
		ManagementUsername:   "root",
		AuthType:   "password",
		Credential: "password123",
	})
	req := httptest.NewRequest("POST", "/api/vms", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()
	handler.CreateVM(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if _, ok := resp["id"]; !ok {
		t.Error("response missing id")
	}
}

func TestGetVMs(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	handler := &VMHandler{DB: testDB}
	testDB.Exec(`INSERT INTO vms (id, name, host, management_username, auth_type, credential, status) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"test-id", "Test VM", "1.2.3.4", "root", "password", "enc", "online")

	req := httptest.NewRequest("GET", "/api/vms", nil)
	w := httptest.NewRecorder()
	handler.GetVMs(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func BenchmarkAddVM(b *testing.B) {
	testDB := setupTestDB(nil)
	defer testDB.Close()

	handler := &VMHandler{DB: testDB}
	os.Setenv("AES_KEY", "this-is-a-32-byte-key-!!-1234567")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reqBody, _ := json.Marshal(CreateVMRequest{
			Name:       "Test VM",
			Host:       "1.2.3.4",
			ManagementUsername:   "root",
			AuthType:   "password",
			Credential: "password123",
		})
		req := httptest.NewRequest("POST", "/api/vms", bytes.NewBuffer(reqBody))
		w := httptest.NewRecorder()
		handler.CreateVM(w, req)
		if w.Code != http.StatusCreated {
			b.Fatalf("Benchmark failed: expected 201, got %d", w.Code)
		}
	}
}

func BenchmarkGetVMs(b *testing.B) {
	testDB := setupTestDB(nil)
	defer testDB.Close()

	handler := &VMHandler{DB: testDB}
	testDB.Exec(`INSERT INTO vms (id, name, host, management_username, auth_type, credential, status) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"test-id", "Test VM", "1.2.3.4", "root", "password", "enc", "online")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/vms", nil)
		w := httptest.NewRecorder()
		handler.GetVMs(w, req)
	}
}
