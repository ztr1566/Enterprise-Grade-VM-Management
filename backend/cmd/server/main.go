package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"backend/internal/api/handlers"
	"backend/internal/api/middleware"
	"backend/internal/audit"
	"backend/internal/db"
	"backend/internal/monitor"
	"backend/internal/ws"
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
	migration002, err := os.ReadFile("internal/db/migrations/002_ssh_keys.sql")
	if err != nil {
		audit.Logger.Fatal("Failed to read migration 002", zap.Error(err))
	}
	// ALTER TABLE errors are expected on re-run (column already exists) — ignore them
	for _, stmt := range splitSQL(string(migration002)) {
		if _, err := sqliteDB.Exec(stmt); err != nil {
			audit.Logger.Warn("Migration 002 stmt skipped (likely already applied)",
				zap.String("stmt", stmt[:min(60, len(stmt))]), zap.Error(err))
		}
	}
	audit.Logger.Info("Migration 002 applied")
 
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

	// 4. Initialize Handlers
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

	// 5b. Start background VM status pinger
	monitor.StartPinger(sqliteDB)

	// 6. Start blocking Server
	port := "8080"
	audit.Logger.Info("Server is starting and listening", zap.String("port", port))

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		audit.Logger.Fatal("Server encountered a fatal error", zap.Error(err))
	}
}

// splitSQL splits a SQL file into individual statements for safe execution.
func splitSQL(sql string) []string {
	var stmts []string
	for _, s := range strings.Split(sql, ";") {
		s = strings.TrimSpace(s)
		if s != "" && !strings.HasPrefix(s, "--") {
			stmts = append(stmts, s)
		}
	}
	return stmts
}

// min returns the smaller of two ints (Go 1.21+ has this built-in, kept for compat).
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
