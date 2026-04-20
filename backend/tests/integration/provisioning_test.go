package integration

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"io"
	"log"
	"net"
	"os"
	"testing"

	"backend/internal/agent/bootstrap"
	"backend/internal/api/grpc/telemetry"
	"backend/internal/audit"
	"backend/internal/ca"
	"backend/internal/crypto"
	"backend/internal/models"
	"backend/internal/ssh"

	_ "github.com/mattn/go-sqlite3"
	gossh "golang.org/x/crypto/ssh"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/test/bufconn"
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
		"../../internal/db/migrations/006_telemetry.sql",
		"../../internal/db/migrations/007_agent_tokens.sql",
		"../../internal/db/migrations/008_agent_certificates.sql",
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
	
	_, err := db.Exec("INSERT INTO vms (id, name, host, management_username, auth_type, credential, status) VALUES (?, ?, ?, ?, ?, ?, ?)",
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
}

func TestZeroTrustTelemetry(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// 1. Setup Backend CA and gRPC Server
	dataDir := t.TempDir()
	caInst, err := ca.LoadOrCreateCA(dataDir)
	if err != nil {
		t.Fatalf("failed to setup CA: %v", err)
	}

	// Create server certificates
	serverCertPEM, serverKeyPEM, err := caInst.GenerateServerCertificate(dataDir, "localhost")
	if err != nil {
		t.Fatalf("failed to generate server cert: %v", err)
	}
	serverCert, _ := tls.X509KeyPair(serverCertPEM, serverKeyPEM)

	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(caInst.CertBytes)

	serverTLS := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.VerifyClientCertIfGiven,
		ClientCAs:    certPool,
	}

	lis := bufconn.Listen(1024 * 1024)
	s := grpc.NewServer(grpc.Creds(credentials.NewTLS(serverTLS)))
	telemetry.RegisterAgentIdentityServer(s, telemetry.NewIdentityHandler(caInst, db))
	telemetry.RegisterTelemetryIngestionServer(s, telemetry.NewTelemetryHandler(db))
	
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Printf("Server exited with error: %v", err)
		}
	}()
	defer s.Stop()

	bufDialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	// 2. Simulate Agent Bootstrap (CSR Flow)
	vmID := "agent-007"
	key, _ := bootstrap.GenerateKey()
	csrPEM, _ := bootstrap.GenerateCSR(key, vmID)

	// Dial for bootstrap (VerifyClientCertIfGiven allows this without client cert)
	conn, err := grpc.DialContext(context.Background(), "bufnet", 
		grpc.WithContextDialer(bufDialer), 
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})))
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	
	idClient := telemetry.NewAgentIdentityClient(conn)
	resp, err := idClient.SignCSR(context.Background(), &telemetry.CSRRequest{
		VmId:   vmID,
		CsrPem: csrPEM,
	})
	if err != nil {
		t.Fatalf("SignCSR failed: %v", err)
	}
	conn.Close()

	// 3. Simulate Agent Streaming (mTLS Flow)
	keyBytes, _ := x509.MarshalECPrivateKey(key)
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes})

	cert, err := tls.X509KeyPair(resp.CertificatePem, keyPEM)
	if err != nil {
		t.Fatalf("Failed to create key pair: %v", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      certPool,
		ServerName:   "localhost",
	}

	conn, err = grpc.DialContext(context.Background(), "bufnet",
		grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	if err != nil {
		t.Fatalf("Failed to dial mTLS: %v", err)
	}
	defer conn.Close()

	ingestClient := telemetry.NewTelemetryIngestionClient(conn)
	stream, err := ingestClient.StreamMetricsBatch(context.Background())
	if err != nil {
		t.Fatalf("StreamMetricsBatch failed: %v", err)
	}

	batch := &telemetry.MetricsBatch{
		VmId: vmID,
		Samples: []*telemetry.MetricSample{
			{
				Timestamp:        123456789,
				CpuUsagePercent:  42.5,
				MemoryUsedBytes:  1024,
				MemoryTotalBytes: 2048,
			},
		},
	}

	if err := stream.Send(batch); err != nil {
		t.Fatalf("Send failed: %v", err)
	}
	
	ack, err := stream.CloseAndRecv()
	if err != nil {
		t.Fatalf("CloseAndRecv failed: %v", err)
	}
	if ack.ProcessedCount != 1 {
		t.Errorf("Expected 1 processed sample, got %d", ack.ProcessedCount)
	}

	// 4. Verify DB Ingestion
	var cpu float64
	err = db.QueryRow("SELECT cpu_usage FROM metrics WHERE vm_id = ?", vmID).Scan(&cpu)
	if err != nil {
		t.Fatalf("DB query failed: %v", err)
	}
	if cpu != 42.5 {
		t.Errorf("Expected CPU 42.5, got %f", cpu)
	}
}
