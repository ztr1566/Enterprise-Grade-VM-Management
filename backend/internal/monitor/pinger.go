package monitor

import (
	"database/sql"
	"time"

	"go.uber.org/zap"
	"backend/internal/audit"
	sshclient "backend/internal/ssh"
)

const (
	pingInterval = 30 * time.Second
	dialTimeout  = 3 * time.Second
)

// StartPinger launches a background goroutine that probes all VMs via TCP every 30 seconds
// and updates their status to 'online' or 'offline' in the database.
func StartPinger(db *sql.DB) {
	go func() {
		audit.Logger.Info("VM status pinger started", zap.Duration("interval", pingInterval))
		// Ping once immediately on startup, then on the tick interval
		pingAll(db)

		ticker := time.NewTicker(pingInterval)
		defer ticker.Stop()
		for range ticker.C {
			pingAll(db)
		}
	}()
}

// pingAll fetches all VMs and probes each one concurrently.
func pingAll(db *sql.DB) {
	rows, err := db.Query(`SELECT id FROM vms`)
	if err != nil {
		audit.Logger.Error("Pinger: failed to query VMs", zap.Error(err))
		return
	}
	defer rows.Close()

	var vmIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err == nil {
			vmIDs = append(vmIDs, id)
		}
	}

	// Probe concurrently to avoid blocking the tick on slow hosts
	for _, id := range vmIDs {
		go probeVM(db, id)
	}
}

// probeVM attempts an SSH connection using the management_username. Updates the DB status accordingly.
func probeVM(db *sql.DB, vmID string) {
	session, err := sshclient.NewSession(db, vmID)
	
	var status string
	if err == nil {
		session.Close()
		status = "online"
	} else {
		status = "offline"
	}

	if _, dbErr := db.Exec(`UPDATE vms SET status = ? WHERE id = ?`, status, vmID); dbErr != nil {
		audit.Logger.Error("Pinger: failed to update VM status", zap.String("vmID", vmID), zap.Error(dbErr))
		return
	}

	audit.Logger.Debug("Pinger: probed VM", zap.String("vmID", vmID), zap.String("status", status))
}
