package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"backend/internal/api"
	"backend/internal/audit"
	"backend/internal/models"
	"backend/internal/monitor"
	sshclient "backend/internal/ssh"
)

type ServiceHandler struct {
	DB *sql.DB
}

// ServiceStatus represents a systemd unit's live state.
type ServiceStatus struct {
	Name   string `json:"name"`
	Load   string `json:"load"`   // loaded | not-found | masked
	Active string `json:"active"` // active | inactive | activating | failed | OFFLINE
	Sub    string `json:"sub"`    // running | exited | dead | failed …
}

var whitelistNames = []string{"ssh", "docker", "nginx", "mysql", "redis", "adguard", "firewall", "ufw", "cron", "systemd-journald"}

// GetServices fetches live service state from the remote VM using
// systemctl list-units and filters for specific services.
func (h *ServiceHandler) GetServices(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		api.WriteError(w, http.StatusBadRequest, "Missing VM ID")
		return
	}

	services, err := monitor.FetchServices(h.DB, id)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "Failed to fetch services: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(services)
}

type ActionRequest struct {
	Action string `json:"action"`
}

// PerformAction executes a systemctl action on a specific service and audits the event.
func (h *ServiceHandler) PerformAction(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	name := r.PathValue("name")

	var req ActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	userID, _ := r.Context().Value("user_id").(string)

	out, cmdErr := sshclient.RunSystemdCommand(h.DB, id, name, req.Action)

	// Audit: SERVICE_ACTION_EXECUTED
	details := fmt.Sprintf("user=%s vm=%s service=%s action=%s success=%v", userID, id, name, req.Action, cmdErr == nil)
	if err := models.CreateAuditLog(h.DB, "SERVICE_ACTION_EXECUTED", details); err != nil {
		audit.Logger.Error("Failed to write audit log for service action", zap.Error(err))
	}
	audit.Logger.Info("Service action executed",
		zap.String("vmID", id), zap.String("service", name),
		zap.String("action", req.Action), zap.String("userID", userID))

	w.Header().Set("Content-Type", "application/json")
	// systemctl stop/start/restart return non-zero on failure but `status` returns
	// non-zero when the service is inactive — treat that as a success.
	if cmdErr != nil && req.Action != "status" {
		json.NewEncoder(w).Encode(map[string]string{
			"status": "error",
			"output": string(out),
			"error":  cmdErr.Error(),
		})
		return
	}
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
		"output": string(out),
	})
}
