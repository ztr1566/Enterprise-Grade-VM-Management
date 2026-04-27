package main

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"backend/internal/api/grpc/interceptors"
	"backend/internal/api/grpc/telemetry"
	"backend/internal/api/handlers"
	"backend/internal/api/middleware"
	"backend/internal/audit"
	"backend/internal/ca"
	"backend/internal/db"
	"backend/internal/models"
	"backend/internal/monitor"
	"backend/internal/ws"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	// 0. Load environment variables from .env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	// 1. Initialize Audit Logger (Zap)
	audit.InitLogger()
	defer audit.Logger.Sync()
	audit.Logger.Info("Initializing Web-Based VM Management Platform...")

	// 2. Initialize SQLite Database
	sqliteDB, err := db.InitDB("vm-manager.db")
	if err != nil {
		audit.Logger.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer sqliteDB.Close()

	// 3. Run Migrations
	migration, err := os.ReadFile("internal/db/migrations/001_init.sql")
	if err != nil {
		audit.Logger.Fatal("Failed to read migration file", zap.Error(err))
	}
	if _, err := sqliteDB.Exec(string(migration)); err != nil {
		audit.Logger.Fatal("Failed to run migrations", zap.Error(err))
	}
	audit.Logger.Info("Database migrations applied successfully")

	// 3b. Run Migration 002: SSH Key Vault
	var count2 int
	_ = sqliteDB.QueryRow("SELECT count(*) FROM pragma_table_info('vms') WHERE name='key_id'").Scan(&count2)
	if count2 == 0 {
		migration002, err := os.ReadFile("internal/db/migrations/002_ssh_keys.sql")
		if err != nil {
			audit.Logger.Fatal("Failed to read migration 002", zap.Error(err))
		}
		for _, stmt := range splitSQL(string(migration002)) {
			if _, err := sqliteDB.Exec(stmt); err != nil {
				audit.Logger.Fatal("Failed to run migration 002", zap.Error(err))
			}
		}
		audit.Logger.Info("Migration 002 applied")
	} else {
		audit.Logger.Info("Migration 002 already applied")
	}
 
	// 3c. Run Migration 003: Provisioning Records
	migration003, err := os.ReadFile("internal/db/migrations/003_provisioning.sql")
	if err != nil {
		audit.Logger.Fatal("Failed to read migration 003", zap.Error(err))
	}
	if _, err := sqliteDB.Exec(string(migration003)); err != nil {
		audit.Logger.Fatal("Failed to run migration 003", zap.Error(err))
	}
	audit.Logger.Info("Migration 003 applied")
 
	// 3d. Run Migration 004: VM Access
	migration004, err := os.ReadFile("internal/db/migrations/004_vm_access.sql")
	if err != nil {
		audit.Logger.Fatal("Failed to read migration 004", zap.Error(err))
	}
	if _, err := sqliteDB.Exec(string(migration004)); err != nil {
		audit.Logger.Fatal("Failed to run migration 004", zap.Error(err))
	}
	audit.Logger.Info("Migration 004 applied")
 
	// 3e. Run Migration 005: Identity Refactor (Idempotent check)
	var count int
	_ = sqliteDB.QueryRow("SELECT count(*) FROM pragma_table_info('vms') WHERE name='username'").Scan(&count)
	if count > 0 {
		migration005, err := os.ReadFile("internal/db/migrations/005_identity_refactor.sql")
		if err != nil {
			audit.Logger.Fatal("Failed to read migration 005", zap.Error(err))
		}
		if _, err := sqliteDB.Exec(string(migration005)); err != nil {
			audit.Logger.Fatal("Failed to run migration 005", zap.Error(err))
		}
		audit.Logger.Info("Migration 005 applied")
	} else {
		audit.Logger.Info("Migration 005 already applied (column renamed)")
	}

	// 3f. Run Migration 006: Telemetry Tables
	migration006, err := os.ReadFile("internal/db/migrations/006_telemetry.sql")
	if err != nil {
		audit.Logger.Fatal("Failed to read migration 006", zap.Error(err))
	}
	if _, err := sqliteDB.Exec(string(migration006)); err != nil {
		audit.Logger.Fatal("Failed to run migration 006", zap.Error(err))
	}
	audit.Logger.Info("Migration 006 applied")

	// 3g. Run Migration 007: Agent Tokens (OTT)
	migration007, err := os.ReadFile("internal/db/migrations/007_agent_tokens.sql")
	if err != nil {
		audit.Logger.Fatal("Failed to read migration 007", zap.Error(err))
	}
	if _, err := sqliteDB.Exec(string(migration007)); err != nil {
		audit.Logger.Fatal("Failed to run migration 007", zap.Error(err))
	}
	audit.Logger.Info("Migration 007 applied")

	// 3h. Run Migration 008: Agent Certificates
	migration008, err := os.ReadFile("internal/db/migrations/008_agent_certificates.sql")
	if err != nil {
		audit.Logger.Fatal("Failed to read migration 008", zap.Error(err))
	}
	if _, err := sqliteDB.Exec(string(migration008)); err != nil {
		audit.Logger.Fatal("Failed to run migration 008", zap.Error(err))
	}
	audit.Logger.Info("Migration 008 applied")

	// 3i. Run Migration 009: VM Last Seen
	var count9 int
	_ = sqliteDB.QueryRow("SELECT count(*) FROM pragma_table_info('vms') WHERE name='last_seen_at'").Scan(&count9)
	if count9 == 0 {
		migration009, err := os.ReadFile("internal/db/migrations/009_vm_last_seen.sql")
		if err != nil {
			audit.Logger.Fatal("Failed to read migration 009", zap.Error(err))
		}
		if _, err := sqliteDB.Exec(string(migration009)); err != nil {
			audit.Logger.Fatal("Failed to run migration 009", zap.Error(err))
		}
		audit.Logger.Info("Migration 009 applied")
	} else {
		audit.Logger.Info("Migration 009 already applied")
	}
 
	// 4. CLI Subcommands
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "provision-token":
			provisionToken(sqliteDB)
			return
		case "revoke-agent":
			revokeAgent(sqliteDB)
			return
		}
	}
 
	// 5. Initialize Handlers
	authHandler := &handlers.AuthHandler{DB: sqliteDB}
	vmHandler := handlers.NewVMHandler(sqliteDB, audit.Logger)
	keyHandler := &handlers.KeyHandler{DB: sqliteDB}
	serviceHandler := &handlers.ServiceHandler{DB: sqliteDB}
	logHandler := &monitor.LogHandler{DB: sqliteDB}
	terminalHandler := &ws.TerminalHandler{DB: sqliteDB}
	provisioningHandler := &handlers.ProvisioningHandler{DB: sqliteDB}
	auditHandler := &handlers.AuditHandler{DB: sqliteDB}
	dashboardHandler := &handlers.DashboardHandler{DB: sqliteDB}
 
	// 5. Setup Router
	mux := http.NewServeMux()
 
	// Public Routes — login is rate-limited: max 5 attempts / minute / IP (T023)
	mux.Handle("POST /api/auth/login",
		middleware.LoginRateLimitMiddleware(http.HandlerFunc(authHandler.LoginHandler)))
 
	// Protected VM REST Routes
	mux.Handle("GET /api/vms", middleware.AuthMiddleware(http.HandlerFunc(vmHandler.GetVMs)))
	mux.Handle("POST /api/vms", middleware.AuthMiddleware(http.HandlerFunc(vmHandler.CreateVM)))
	mux.Handle("GET /api/vms/{id}", middleware.AuthMiddleware(http.HandlerFunc(vmHandler.GetVM)))
	mux.Handle("PUT /api/vms/{id}", middleware.AuthMiddleware(http.HandlerFunc(vmHandler.UpdateVM)))
	mux.Handle("DELETE /api/vms/{id}", middleware.AuthMiddleware(http.HandlerFunc(vmHandler.DeleteVM)))
	mux.Handle("GET /api/vms/{id}/stats", middleware.AuthMiddleware(http.HandlerFunc(vmHandler.GetVMStats)))
	mux.Handle("POST /api/vms/{id}/power", middleware.AuthMiddleware(http.HandlerFunc(vmHandler.PowerAction)))
	mux.Handle("GET /api/vms/{id}/users", middleware.AuthMiddleware(http.HandlerFunc(vmHandler.GetVMUsers)))
	mux.Handle("POST /api/vms/{id}/users", middleware.AuthMiddleware(http.HandlerFunc(vmHandler.AddVMUser)))
	mux.Handle("DELETE /api/vms/{id}/access/{username}", middleware.AuthMiddleware(http.HandlerFunc(vmHandler.RevokeVMUser)))
	mux.Handle("POST /api/vms/{id}/processes/{pid}/signal", middleware.AuthMiddleware(http.HandlerFunc(vmHandler.ProcessAction)))

	// Protected Dashboard Stats Route
	mux.Handle("GET /api/dashboard/stats", middleware.AuthMiddleware(http.HandlerFunc(dashboardHandler.GetDashboardStats)))
 
	// Protected Provisioning Routes
	mux.Handle("GET /api/vms/{id}/provisioning", middleware.AuthMiddleware(http.HandlerFunc(provisioningHandler.GetProvisioningStatus)))
	mux.Handle("POST /api/vms/{id}/provision", middleware.AuthMiddleware(http.HandlerFunc(provisioningHandler.TriggerProvisioning)))

	// Protected Service Routes
	mux.Handle("GET /api/vms/{id}/services", middleware.AuthMiddleware(http.HandlerFunc(serviceHandler.GetServices)))
	mux.Handle("POST /api/vms/{id}/services/{name}/action", middleware.AuthMiddleware(http.HandlerFunc(serviceHandler.PerformAction)))

	// Protected Security Routes (Phase 7 — Firewall & SELinux)
	securityHandler := &handlers.SecurityHandler{DB: sqliteDB}
	mux.Handle("GET /api/vms/{id}/security", middleware.AuthMiddleware(http.HandlerFunc(securityHandler.GetSecurityStatus)))
	mux.Handle("POST /api/vms/{id}/firewall/rules", middleware.AuthMiddleware(http.HandlerFunc(securityHandler.ManageFirewallRule)))
	mux.Handle("POST /api/vms/{id}/selinux", middleware.AuthMiddleware(http.HandlerFunc(securityHandler.SetSELinuxMode)))
	mux.Handle("POST /api/vms/{id}/security/install", middleware.AuthMiddleware(http.HandlerFunc(securityHandler.InstallSecurityComponent)))

	// Protected SSH Key Vault Routes
	mux.Handle("GET /api/keys", middleware.AuthMiddleware(http.HandlerFunc(keyHandler.GetKeys)))
	mux.Handle("POST /api/keys", middleware.AuthMiddleware(http.HandlerFunc(keyHandler.CreateKey)))
	mux.Handle("DELETE /api/keys/{id}", middleware.AuthMiddleware(http.HandlerFunc(keyHandler.DeleteKey)))

	// Protected Audit Log Routes (T017 - Phase 4)
	mux.Handle("GET /api/audit", middleware.AuthMiddleware(http.HandlerFunc(auditHandler.GetAuditLogs)))

	// Protected WebSocket Terminal Route (T018/T019 – Phase 4)
	mux.Handle("GET /api/vms/{id}/terminal", middleware.AuthMiddleware(http.HandlerFunc(terminalHandler.ServeTerminal)))

	// Protected WebSocket Log Route (T035/T036 – Phase 8)
	mux.Handle("GET /api/vms/{id}/logs/{service}", middleware.AuthMiddleware(http.HandlerFunc(logHandler.ServeLogs)))

	// 5b. Start background VM status checker (marks as offline after 30s of silence)
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		for range ticker.C {
			// Mark as offline if no telemetry in last 60 seconds (grace period)
			_, err := sqliteDB.Exec(`
				UPDATE vms 
				SET status = 'offline' 
				WHERE status = 'online' 
				AND (last_seen_at IS NULL OR datetime(last_seen_at) < datetime('now', '-60 seconds'))
			`)
			if err != nil {
				audit.Logger.Error("Background status checker failed", zap.Error(err))
			}
		}
	}()

	// 5c. Start gRPC Server for Telemetry and Identity
	go func() {
		caInst, err := ca.LoadOrCreateCA("data/ca")
		if err != nil {
			audit.Logger.Fatal("Failed to load/create CA", zap.Error(err))
		}

		// Generate server cert for localhost/internal use
		serverCertPEM, serverKeyPEM, err := caInst.GenerateServerCertificate("data/ca", "localhost")
		if err != nil {
			audit.Logger.Fatal("Failed to generate server certificate", zap.Error(err))
		}

		serverCert, err := tls.X509KeyPair(serverCertPEM, serverKeyPEM)
		if err != nil {
			audit.Logger.Fatal("Failed to load server key pair", zap.Error(err))
		}

		certPool := x509.NewCertPool()
		certPool.AppendCertsFromPEM(caInst.CertBytes)

		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{serverCert},
			ClientAuth:   tls.VerifyClientCertIfGiven, // Allow unauthenticated CSR bootstrap
			ClientCAs:    certPool,
			MinVersion:   tls.VersionTLS13,
		}

		grpcServer := grpc.NewServer(
			grpc.Creds(credentials.NewTLS(tlsConfig)),
			grpc.UnaryInterceptor(interceptors.CRLInterceptor(sqliteDB)),
			grpc.ChainStreamInterceptor(
				interceptors.CRLStreamInterceptor(sqliteDB),
				interceptors.PayloadSizeInterceptor(),
			),
		)

		// Register Identity Handler
		telemetry.RegisterAgentIdentityServer(grpcServer, telemetry.NewIdentityHandler(caInst, sqliteDB))
		// Register Telemetry Ingestion Handler
		telemetry.RegisterTelemetryIngestionServer(grpcServer, telemetry.NewTelemetryHandler(sqliteDB))

		lis, err := net.Listen("tcp", ":50051")
		if err != nil {
			audit.Logger.Fatal("gRPC failed to listen", zap.Error(err))
		}

		audit.Logger.Info("gRPC server listening", zap.String("port", "50051"))
		if err := grpcServer.Serve(lis); err != nil {
			audit.Logger.Fatal("gRPC server failed", zap.Error(err))
		}
	}()

	// 6. Start blocking Server
	port := "8080"
	audit.Logger.Info("Server is starting and listening", zap.String("port", port))

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		audit.Logger.Fatal("Server encountered a fatal error", zap.Error(err))
	}
}

// splitSQL splits a SQL file into individual statements for safe execution.
// It strips comment-only lines within each segment before filtering, so a
// statement preceded by a "-- comment" line is not accidentally dropped.
func splitSQL(content string) []string {
	var stmts []string
	for _, segment := range strings.Split(content, ";") {
		// Strip comment-only lines from the segment, then trim whitespace.
		var lines []string
		for _, line := range strings.Split(segment, "\n") {
			if !strings.HasPrefix(strings.TrimSpace(line), "--") {
				lines = append(lines, line)
			}
		}
		s := strings.TrimSpace(strings.Join(lines, "\n"))
		if s != "" {
			stmts = append(stmts, s)
		}
	}
	return stmts
}

func provisionToken(db *sql.DB) {
	fs := flag.NewFlagSet("provision-token", flag.ExitOnError)
	machineID := fs.String("machine-id", "", "ID of the machine to provision")
	fs.Parse(os.Args[2:])

	if *machineID == "" {
		fmt.Println("Error: --machine-id is required")
		os.Exit(1)
	}

	token := uuid.New().String()
	hash := sha256.Sum256([]byte(token))
	hashStr := hex.EncodeToString(hash[:])

	agentToken := models.AgentToken{
		ID:        uuid.New().String(),
		MachineID: *machineID,
		TokenHash: hashStr,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}

	if err := models.CreateAgentToken(db, agentToken); err != nil {
		fmt.Printf("Error: failed to persist token: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Token: %s\n(Valid for 10 minutes for machine: %s)\n", token, *machineID)
}

func revokeAgent(db *sql.DB) {
	fs := flag.NewFlagSet("revoke-agent", flag.ExitOnError)
	machineID := fs.String("machine-id", "", "ID of the machine to revoke")
	fs.Parse(os.Args[2:])

	if *machineID == "" {
		fmt.Println("Error: --machine-id is required")
		os.Exit(1)
	}

	if err := models.RevokeAgentCertificates(db, *machineID); err != nil {
		fmt.Printf("Error: failed to revoke agent: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully revoked all certificates for machine: %s\n", *machineID)
}

