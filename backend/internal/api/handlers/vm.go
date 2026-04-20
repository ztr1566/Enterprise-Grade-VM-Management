package handlers

import (
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"backend/internal/db"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"backend/internal/api"
	"backend/internal/audit"
	"backend/internal/crypto"
	"backend/internal/models"
	"backend/internal/monitor"
	sshclient "backend/internal/ssh"
)

// VMHandler encapsulates VM-related API endpoints.
type VMHandler struct {
	DB     *sql.DB
	Logger *zap.Logger
}

// NewVMHandler creates a new VMHandler with injected dependencies.
func NewVMHandler(db *sql.DB, logger *zap.Logger) *VMHandler {
	return &VMHandler{
		DB:     db,
		Logger: logger,
	}
}

// CreateVMRequest defines the expected payload for creating a VM.
type CreateVMRequest struct {
	Name               string   `json:"name"`
	Host               string   `json:"host"`
	ManagementUsername string   `json:"management_username"`
	AuthType           string   `json:"auth_type"`
	Credential         string   `json:"credential"`
	KeyID              string   `json:"key_id"` // Optional: ID from ssh_keys vault
	Tags               []string `json:"tags"`
}

// UpdateVMRequest defines the payload for editing a VM (credential is optional).
type UpdateVMRequest struct {
	Name               string   `json:"name"`
	Host               string   `json:"host"`
	ManagementUsername string   `json:"management_username"`
	AuthType           string   `json:"auth_type"`
	Credential         string   `json:"credential"` // If empty, credential is unchanged
	KeyID              string   `json:"key_id"`     // If set, vault key is used instead of credential
	Tags               []string `json:"tags"`
}

// GetVMs handles the retrieval of the entire VM inventory.
func (h *VMHandler) GetVMs(w http.ResponseWriter, r *http.Request) {
	vms, err := models.GetVMs(h.DB)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "Failed to fetch VM inventory")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vms)
}

// CreateVM handles the creation of a new VM record with encrypted credentials.
func (h *VMHandler) CreateVM(w http.ResponseWriter, r *http.Request) {
	var req CreateVMRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		audit.Logger.Error("Failed to decode CreateVM request", zap.Error(err))
		api.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Credential is required UNLESS a vault key_id is provided
	if req.Credential == "" && req.KeyID == "" {
		audit.Logger.Warn("Validation failed: neither credential nor key_id provided")
		api.WriteError(w, http.StatusBadRequest, "Either a credential or a vault key_id is required")
		return
	}

	// Only encrypt credential when provided; key_id VMs use vault auth
	encryptedCredential := ""
	if req.Credential != "" {
		aesKey := crypto.MustGetAESKey()
		audit.Logger.Info("Attempting to encrypt", zap.String("input_len", fmt.Sprintf("%d", len(req.Credential))))
		encrypted, err := crypto.Encrypt([]byte(req.Credential), aesKey)
		if err != nil {
			audit.Logger.Error("Encryption failure", zap.Error(err))
			api.WriteError(w, http.StatusInternalServerError, "Failed to encrypt credentials")
			return
		}
		audit.Logger.Info("Encryption result", zap.String("hex", hex.EncodeToString(encrypted)))
		encryptedCredential = base64.StdEncoding.EncodeToString(encrypted)
	}

	vm := models.VM{
		ID:                 uuid.New().String(),
		Name:               req.Name,
		Host:               req.Host,
		ManagementUsername: req.ManagementUsername,
		AuthType:           req.AuthType,
		Credential:         encryptedCredential,
		KeyID:              req.KeyID,
		Tags:               req.Tags,
		Status:             "offline",
	}

	if err := models.CreateVM(h.DB, vm); err != nil {
		audit.Logger.Error("DB persistence failure", zap.Error(err))
		api.WriteError(w, http.StatusInternalServerError, "Failed to persist VM record")
		return
	}
 
	// Create provisioning record
	if err := models.CreateProvisioningRecord(h.DB, vm.ID); err != nil {
		audit.Logger.Error("Failed to create provisioning record", zap.String("vmID", vm.ID), zap.Error(err))
		// We continue anyway, as the VM is already created.
	}
 
	// Launch async provisioning
	go func() {
		if err := sshclient.ProvisionVM(h.DB, vm.ID); err != nil {
			audit.Logger.Error("Async provisioning failed", zap.String("vmID", vm.ID), zap.Error(err))
		}
	}()
 
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
 
	// Return VM + status (anonymous struct to include provisioning_status)
	response := struct {
		models.VM
		ProvisioningStatus string `json:"provisioning_status"`
	}{
		VM:                 vm,
		ProvisioningStatus: "pending",
	}
	json.NewEncoder(w).Encode(response)
}

// UpdateVM handles full or partial update of a VM record.
func (h *VMHandler) UpdateVM(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		api.WriteError(w, http.StatusBadRequest, "Missing VM ID")
		return
	}

	var req UpdateVMRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		audit.Logger.Error("Failed to decode UpdateVM request", zap.Error(err))
		api.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// If credential is provided, encrypt it
	encryptedCredential := ""
	if req.Credential != "" {
		aesKey := crypto.MustGetAESKey()
		encrypted, err := crypto.Encrypt([]byte(req.Credential), aesKey)
		if err != nil {
			api.WriteError(w, http.StatusInternalServerError, "Failed to encrypt credentials")
			return
		}
		encryptedCredential = base64.StdEncoding.EncodeToString(encrypted)
	}

	// Update record in DB
	err := models.UpdateVM(h.DB, id, req.Name, req.Host, req.ManagementUsername, req.AuthType, encryptedCredential, req.KeyID, req.Tags)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "Failed to update VM record")
		return
	}

	audit.Logger.Info("VM updated", zap.String("vmID", id))
	w.WriteHeader(http.StatusNoContent)
}

// DeleteVM handles the deletion of a VM and its associated data.
func (h *VMHandler) DeleteVM(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := models.DeleteVM(h.DB, id); err != nil {
		api.WriteError(w, http.StatusInternalServerError, "Failed to delete VM")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GetVM retrieves a single VM by ID.
func (h *VMHandler) GetVM(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	vm, err := models.GetVMByID(h.DB, id)
	if err != nil {
		api.WriteError(w, http.StatusNotFound, "VM not found")
		return
	}
	json.NewEncoder(w).Encode(vm)
}



// GetVMStats returns the latest agent-pushed CPU/RAM/Disk metrics for a VM from the database.
func (h *VMHandler) GetVMStats(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		api.WriteError(w, http.StatusBadRequest, "Missing VM ID")
		return
	}

	// Phase 6: Fetch latest metrics from DB (agent-pushed) instead of polling via SSH
	m, err := db.GetLatestMetrics(h.DB, id)
	if err != nil {
		if err == sql.ErrNoRows {
			api.WriteError(w, http.StatusNotFound, "No telemetry data available for this VM")
		} else {
			audit.Logger.Error("Failed to fetch VM metrics from DB", zap.Error(err), zap.String("vmID", id))
			api.WriteError(w, http.StatusInternalServerError, "Failed to fetch VM metrics")
		}
		return
	}

	// Map DB metrics to the expected API response
	stats := &monitor.VMStats{
		CPU:  m.CpuUsagePercent,
		RAM: func() float64 {
			if m.MemoryTotalBytes == 0 {
				return 0
			}
			return float64(m.MemoryUsedBytes) / float64(m.MemoryTotalBytes) * 100
		}(),
		Disk: m.DiskUsagePercent,
		Network: &monitor.NetworkMetrics{
			RxBytesSec: 0, // Rate calculation requires two samples, for now return 0 or implement in DB
			TxBytesSec: 0,
		},
		Timestamp: time.UnixMilli(m.Timestamp),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// GetVMUsers retrieves the list of OS usernames permitted for a specific VM.
func (h *VMHandler) GetVMUsers(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	users, err := models.GetUsersByVM(h.DB, id)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "Failed to fetch VM users")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// AddVMUser grants a specific OS username access to a VM.
func (h *VMHandler) AddVMUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Username == "" {
		api.WriteError(w, http.StatusBadRequest, "Username is required")
		return
	}

	if err := models.AddUserToVM(h.DB, id, req.Username); err != nil {
		api.WriteError(w, http.StatusInternalServerError, "Failed to add user to VM")
		return
	}

	// Trigger async provisioning to update sudoers
	go func() {
		if err := sshclient.ProvisionVM(h.DB, id); err != nil {
			audit.Logger.Error("Async provisioning after user add failed", zap.String("vmID", id), zap.Error(err))
		}
	}()

	w.WriteHeader(http.StatusCreated)
}

// RevokeVMUser removes a specific OS username's access from a VM.
func (h *VMHandler) RevokeVMUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	username := r.PathValue("username")

	if id == "" || username == "" {
		api.WriteError(w, http.StatusBadRequest, "Missing VM ID or username")
		return
	}

	// 1. Remove from DB
	if err := models.RemoveUserFromVM(h.DB, id, username); err != nil {
		api.WriteError(w, http.StatusInternalServerError, "Failed to remove user from DB")
		return
	}

	// 2. Trigger async de-provisioning via SSH
	go func() {
		if err := sshclient.RevokeUserAccess(h.DB, id, username); err != nil {
			audit.Logger.Error("Async de-provisioning failed", zap.String("vmID", id), zap.String("username", username), zap.Error(err))
		}
	}()

	// 3. Log audit event
	models.CreateAuditLog(h.DB, "USER_ACCESS_REVOKED", fmt.Sprintf("Access revoked for user %s on VM %s", username, id))

	w.WriteHeader(http.StatusNoContent)
}

// ProcessAction handles signaling or re-nicing a process on a VM.
func (h *VMHandler) ProcessAction(w http.ResponseWriter, r *http.Request) {
	vmID := r.PathValue("id")
	pid := r.PathValue("pid")

	var req struct {
		Signal string `json:"signal"` // KILL, TERM, NICE
		Value  int    `json:"value"`  // For NICE
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	session, err := sshclient.NewSession(h.DB, vmID)
	if err != nil {
		h.Logger.Error("SSH connection failed", zap.String("vmID", vmID), zap.Error(err))
		http.Error(w, "Failed to connect to VM", http.StatusInternalServerError)
		return
	}
	defer session.Close()

	var cmd string
	switch strings.ToUpper(req.Signal) {
	case "KILL":
		cmd = fmt.Sprintf("sudo kill -9 %s", pid)
	case "TERM":
		cmd = fmt.Sprintf("sudo kill -15 %s", pid)
	case "NICE":
		cmd = fmt.Sprintf("sudo renice %d -p %s", req.Value, pid)
	default:
		http.Error(w, "Invalid signal type", http.StatusBadRequest)
		return
	}

	out, err := session.RunCmd(cmd)
	if err != nil {
		h.Logger.Error("Failed to execute process action", zap.String("vmID", vmID), zap.String("pid", pid), zap.String("signal", req.Signal), zap.Error(err), zap.String("output", string(out)))
		http.Error(w, fmt.Sprintf("Action failed: %s", string(out)), http.StatusInternalServerError)
		return
	}

	h.Logger.Info("Process action executed", zap.String("vmID", vmID), zap.String("pid", pid), zap.String("signal", req.Signal))
	w.WriteHeader(http.StatusOK)
}

type PowerRequest struct {
	Action string `json:"action"`
}

// PowerAction executes a power command (reboot, shutdown, sleep) on the VM.
func (h *VMHandler) PowerAction(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		api.WriteError(w, http.StatusBadRequest, "Missing VM ID")
		return
	}

	var req PowerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	userID, _ := r.Context().Value("user_id").(string)

	err := sshclient.RunPowerAction(h.DB, id, req.Action)

	actionUpper := strings.ToUpper(req.Action)
	auditEvent := fmt.Sprintf("VM_%s_TRIGGERED", actionUpper)
	details := fmt.Sprintf("user=%s vm=%s action=%s success=%v", userID, id, req.Action, err == nil)

	if auditErr := models.CreateAuditLog(h.DB, auditEvent, details); auditErr != nil {
		h.Logger.Error("Failed to write audit log for power action", zap.Error(auditErr))
	}

	if err != nil {
		h.Logger.Error("Power action failed", zap.String("vmID", id), zap.String("action", req.Action), zap.Error(err))
		api.WriteError(w, http.StatusInternalServerError, "Failed to execute power action: "+err.Error())
		return
	}

	h.Logger.Info("Power action executed successfully", zap.String("vmID", id), zap.String("action", req.Action), zap.String("userID", userID))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
		"message": fmt.Sprintf("Power action '%s' triggered successfully", req.Action),
	})
}
