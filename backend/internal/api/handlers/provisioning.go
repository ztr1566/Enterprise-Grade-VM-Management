package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"backend/internal/api"
	"backend/internal/audit"
	"backend/internal/models"
	"backend/internal/ssh"
	"go.uber.org/zap"
)

// ProvisioningHandler handles API requests related to VM provisioning.
type ProvisioningHandler struct {
	DB *sql.DB
}

// GetProvisioningStatus returns the current provisioning record for a VM.
func (h *ProvisioningHandler) GetProvisioningStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		audit.Logger.Error("Missing VM ID in GetProvisioningStatus")
		api.WriteError(w, http.StatusBadRequest, "Missing VM ID")
		return
	}

	record, err := models.GetProvisioningRecord(h.DB, id)
	if err != nil {
		audit.Logger.Error("Failed to fetch provisioning record", zap.Error(err), zap.String("vmID", id))
		api.WriteError(w, http.StatusInternalServerError, "Failed to fetch provisioning record")
		return
	}

	if record == nil {
		audit.Logger.Warn("No provisioning record found for this VM", zap.String("vmID", id))
		api.WriteError(w, http.StatusNotFound, "No provisioning record found for this VM")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(record)
}

// TriggerProvisioning manually starts the provisioning engine for a VM.
func (h *ProvisioningHandler) TriggerProvisioning(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		audit.Logger.Error("Missing VM ID in TriggerProvisioning")
		api.WriteError(w, http.StatusBadRequest, "Missing VM ID")
		return
	}

	// Verify VM exists
	_, err := models.GetVMByID(h.DB, id)
	if err != nil {
		audit.Logger.Error("VM not found or failed to query", zap.Error(err), zap.String("vmID", id))
		api.WriteError(w, http.StatusNotFound, "VM not found")
		return
	}

	// Audit: VM_REPROVISION_REQUESTED
	models.CreateAuditLog(h.DB, "VM_REPROVISION_REQUESTED", "Manual re-provisioning triggered for VM "+id)

	// Ensure provisioning record exists (ignore error if it already exists)
	_ = models.CreateProvisioningRecord(h.DB, id)

	// Update status to pending synchronously
	if err := models.UpdateProvisioningStatus(h.DB, id, "pending", nil, nil); err != nil {
		audit.Logger.Error("Failed to update provisioning status to pending", zap.String("vmID", id), zap.Error(err))
		api.WriteError(w, http.StatusInternalServerError, "Failed to initialize provisioning state")
		return
	}

	// Launch async provisioning
	go func() {
		if err := ssh.ProvisionVM(h.DB, id); err != nil {
			audit.Logger.Error("Async provisioning failed", zap.String("vmID", id), zap.Error(err))
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"vm_id":               id,
		"provisioning_status": "pending",
		"message":             "Provisioning initiated",
	})
}
