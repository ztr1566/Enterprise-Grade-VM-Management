package audit

import (
	"database/sql"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// Event type constants for SSH terminal sessions.
const (
	EventSSHSessionStarted  = "SSH_SESSION_STARTED"
	EventSSHSessionEnded    = "SSH_SESSION_ENDED"
	EventSSHConnectionFailed = "SSH_CONNECTION_FAILED"
)

// LogSSHSessionStarted records a new SSH terminal session to both Zap and the DB.
func LogSSHSessionStarted(db *sql.DB, userID, vmID string) {
	Logger.Info(EventSSHSessionStarted,
		zap.String("user_id", userID),
		zap.String("vm_id", vmID),
	)
	if db != nil {
		writeAuditLog(db, EventSSHSessionStarted,
			fmt.Sprintf("SSH session started by user %s on VM %s", userID, vmID))
	}
}

// LogSSHSessionEnded records session closure with both a human-readable duration
// and the raw seconds to the DB details column, then logs via Zap.
func LogSSHSessionEnded(db *sql.DB, userID, vmID string, duration time.Duration) {
	seconds := int64(duration.Seconds())
	human := duration.Round(time.Second).String()

	Logger.Info(EventSSHSessionEnded,
		zap.String("user_id", userID),
		zap.String("vm_id", vmID),
		zap.Int64("duration_seconds", seconds),
		zap.String("duration_human", human),
	)
	if db != nil {
		details := fmt.Sprintf(
			"SSH session ended | user=%s vm=%s duration=%s (%ds)",
			userID, vmID, human, seconds,
		)
		writeAuditLog(db, EventSSHSessionEnded, details)
	}
}

// LogSSHConnectionFailed records a failed SSH connection attempt with the error detail.
func LogSSHConnectionFailed(db *sql.DB, userID, vmID, reason string) {
	Logger.Error(EventSSHConnectionFailed,
		zap.String("user_id", userID),
		zap.String("vm_id", vmID),
		zap.String("reason", reason),
	)
	if db != nil {
		writeAuditLog(db, EventSSHConnectionFailed,
			fmt.Sprintf("SSH connection failed for user %s on VM %s: %s", userID, vmID, reason))
	}
}

// writeAuditLog is an internal helper that persists an event to audit_log_entries.
func writeAuditLog(db *sql.DB, eventType, details string) {
	if _, err := db.Exec(
		`INSERT INTO audit_log_entries (event_type, details, timestamp) VALUES (?, ?, ?)`,
		eventType, details, time.Now().UTC().Format(time.RFC3339),
	); err != nil {
		// Non-fatal: log the DB error but don't crash the session
		Logger.Error("Failed to persist audit log",
			zap.String("event_type", eventType),
			zap.Error(err),
		)
	}
}
