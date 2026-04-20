package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// InitDB initializes the SQLite database connection and creates tables if they don't exist.
func InitDB(dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Telemetry tables (metrics, logs) are created by migration 006_telemetry.sql

	log.Println("Successfully connected to SQLite database and initialized schema")
	return db, nil
}

// MetricSample is a local representation of telemetry metrics for database storage.
type MetricSample struct {
	Timestamp         int64
	CpuUsagePercent   float64
	MemoryUsedBytes   uint64
	MemoryTotalBytes  uint64
	DiskUsagePercent  float64
	NetworkTxBytes    uint64
	NetworkRxBytes    uint64
}

// InsertMetrics stores a batch of metric samples in the database.
func InsertMetrics(db *sql.DB, vmID string, samples []*MetricSample) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO metrics (vm_id, timestamp, cpu_usage, memory_used, memory_total, disk_usage, network_tx, network_rx)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, s := range samples {
		_, err := stmt.Exec(vmID, s.Timestamp, s.CpuUsagePercent, s.MemoryUsedBytes, s.MemoryTotalBytes, s.DiskUsagePercent, s.NetworkTxBytes, s.NetworkRxBytes)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// LogEntry is a local representation of a log entry for database storage.
type LogEntry struct {
	Timestamp int64
	Severity  string
	Component string
	Message   string
}

// InsertLogs stores a batch of log entries in the database.
func InsertLogs(db *sql.DB, vmID string, entries []*LogEntry) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO logs (vm_id, timestamp, severity, component, message)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, e := range entries {
		_, err := stmt.Exec(vmID, e.Timestamp, e.Severity, e.Component, e.Message)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetLatestMetrics fetches the most recent metric sample for a given VM.
func GetLatestMetrics(db *sql.DB, vmID string) (*MetricSample, error) {
	row := db.QueryRow(`
		SELECT timestamp, cpu_usage, memory_used, memory_total, disk_usage, network_tx, network_rx
		FROM metrics
		WHERE vm_id = ?
		ORDER BY timestamp DESC
		LIMIT 1
	`, vmID)

	var s MetricSample
	err := row.Scan(&s.Timestamp, &s.CpuUsagePercent, &s.MemoryUsedBytes, &s.MemoryTotalBytes, &s.DiskUsagePercent, &s.NetworkTxBytes, &s.NetworkRxBytes)
	if err != nil {
		return nil, err
	}
	return &s, nil
}
