package monitor

import (
	"database/sql"
	"fmt"
)

// ServiceInfo represents a systemd service's state.
type ServiceInfo struct {
	Name        string `json:"name"`
	LoadState   string `json:"load_state"`
	ActiveState string `json:"active_state"`
	SubState    string `json:"sub_state"`
	Description string `json:"description"`
}

// FetchServices is deprecated in Phase 6. Use agent-pushed telemetry.
func FetchServices(db *sql.DB, vmID string) ([]ServiceInfo, error) {
	return nil, fmt.Errorf("SSH-based service polling is deprecated. Use agent telemetry.")
}
