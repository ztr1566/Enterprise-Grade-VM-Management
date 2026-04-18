package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"backend/internal/api"
	"backend/internal/audit"
	"go.uber.org/zap"
)

// DashboardHandler handles the dashboard overview API.
type DashboardHandler struct {
	DB *sql.DB
}

// DashboardStats contains the aggregate metrics for the dashboard hero cards.
type DashboardStats struct {
	TotalVMs    int `json:"total_vms"`
	OnlineVMs   int `json:"online_vms"`
	OfflineVMs  int `json:"offline_vms"`
	Provisioned int `json:"provisioned"`
	Pending     int `json:"pending"`
	Failed      int `json:"failed"`
	AuditEvents int `json:"audit_events"`
}

// GetDashboardStats returns aggregate VM/provisioning/audit counts.
func (h *DashboardHandler) GetDashboardStats(w http.ResponseWriter, r *http.Request) {
	var stats DashboardStats

	err := h.DB.QueryRow(`SELECT count(*) FROM vms`).Scan(&stats.TotalVMs)
	if err != nil {
		audit.Logger.Error("Dashboard: failed to count VMs", zap.Error(err))
		api.WriteError(w, http.StatusInternalServerError, "Failed to fetch dashboard stats")
		return
	}

	_ = h.DB.QueryRow(`SELECT count(*) FROM vms WHERE status = 'online'`).Scan(&stats.OnlineVMs)
	_ = h.DB.QueryRow(`SELECT count(*) FROM vms WHERE status = 'offline'`).Scan(&stats.OfflineVMs)
	_ = h.DB.QueryRow(`SELECT count(*) FROM provisioning_records WHERE status = 'provisioned'`).Scan(&stats.Provisioned)
	_ = h.DB.QueryRow(`SELECT count(*) FROM provisioning_records WHERE status IN ('pending', 'not_provisioned')`).Scan(&stats.Pending)
	_ = h.DB.QueryRow(`SELECT count(*) FROM provisioning_records WHERE status = 'failed'`).Scan(&stats.Failed)
	_ = h.DB.QueryRow(`SELECT count(*) FROM audit_log_entries`).Scan(&stats.AuditEvents)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
