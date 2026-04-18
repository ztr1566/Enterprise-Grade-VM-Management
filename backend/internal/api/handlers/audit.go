package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"backend/internal/api"
	"backend/internal/models"
)

// AuditHandler handles API requests related to audit logs.
type AuditHandler struct {
	DB *sql.DB
}

// GetAuditLogs returns paginated audit logs.
func (h *AuditHandler) GetAuditLogs(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page := 1
	limit := 50

	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
		limit = l
	}

	offset := (page - 1) * limit

	logs, total, err := models.GetAuditLogs(h.DB, limit, offset)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "Failed to fetch audit logs")
		return
	}

	totalPages := (total + limit - 1) / limit

	response := map[string]interface{}{
		"data": logs,
		"meta": map[string]interface{}{
			"current_page": page,
			"per_page":     limit,
			"total_items":  total,
			"total_pages":  totalPages,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
