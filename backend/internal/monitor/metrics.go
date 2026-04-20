package monitor

import (
	"time"
)

// VMStats represents the real-time resource usage of a VM.
type VMStats struct {
	CPU       float64         `json:"cpu"`
	RAM       float64         `json:"ram"`
	Disk      float64         `json:"disk"`
	Network   *NetworkMetrics `json:"network"`
	Processes []ProcessInfo   `json:"processes"`
	Timestamp time.Time       `json:"timestamp"`
}

// ProcessInfo represents a single process running on the VM.
type ProcessInfo struct {
	PID     string `json:"pid"`
	User    string `json:"user"`
	CPU     string `json:"cpu"`
	Mem     string `json:"mem"`
	Command string `json:"command"`
}

// NetworkMetrics contains the calculated rates and connection counts.
type NetworkMetrics struct {
	RxBytesSec        float64 `json:"rx_bytes_sec"`
	TxBytesSec        float64 `json:"tx_bytes_sec"`
	ActiveConnections int     `json:"active_connections"`
}
