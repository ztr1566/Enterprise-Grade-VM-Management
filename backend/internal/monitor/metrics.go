package monitor

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	sshclient "backend/internal/ssh"
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

// GetVMStats establishes an SSH connection, runs commands to gather metrics,
// and parses the output into a VMStats struct.
func GetVMStats(db *sql.DB, vmID string, query string) (*VMStats, error) {
	session, err := sshclient.NewSession(db, vmID)
	if err != nil {
		return nil, fmt.Errorf("ssh connection failed: %w", err)
	}
	defer session.Close()

	// Run all three commands separated by "---" so we can parse them from a single session execution
	cmd := `top -bn1 | grep 'Cpu(s)' | sed 's/.*, *\([0-9.]*\)%* id.*/\1/' | awk '{print 100 - $1}' && echo "---" && free -m | awk '/Mem:/ { printf("%.2f", $3/$2*100) }' && echo "---" && df -h / | awk 'NR==2 {print $5}' | sed 's/%//'`
	out, err := session.RunCmd(cmd)
	if err != nil {
		return nil, fmt.Errorf("command execution failed: %w", err)
	}

	parts := strings.Split(strings.TrimSpace(string(out)), "---")
	if len(parts) != 3 {
		return nil, fmt.Errorf("unexpected output format: %s", string(out))
	}

	cpu, _ := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	ram, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	disk, _ := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)

	// Fetch network metrics (independently, as it uses its own sample tracking)
	netMetrics, _ := GetNetworkMetrics(db, vmID)

	// Fetch top processes (with optional search)
	processes, _ := FetchProcesses(db, vmID, query)

	return &VMStats{
		CPU:       cpu,
		RAM:       ram,
		Disk:      disk,
		Network:   netMetrics,
		Processes: processes,
		Timestamp: time.Now().UTC(),
	}, nil
}

// FetchProcesses retrieves the top processes by CPU usage from the VM via SSH.
func FetchProcesses(db *sql.DB, vmID string, query string) ([]ProcessInfo, error) {
	session, err := sshclient.NewSession(db, vmID)
	if err != nil {
		return nil, err
	}
	defer session.Close()

	// Base ps command (using pcpu and pmem for broader compatibility)
	psCmd := `ps -eo pid,user,pcpu,pmem,comm --sort=-pcpu`
	
	var cmd string
	if query == "" {
		cmd = fmt.Sprintf("%s | head -21 | awk 'NR>1 {print $1 \"---\" $2 \"---\" $3 \"---\" $4 \"---\" $5}'", psCmd)
	} else {
		// Basic sanitization: remove single quotes and backslashes
		safeQuery := strings.ReplaceAll(query, "'", "")
		safeQuery = strings.ReplaceAll(safeQuery, "\\", "")
		// Bind search to COMMAND field (5th column) with case-insensitive tolower
		cmd = fmt.Sprintf("%s | awk -v q='%s' 'NR>1 && (tolower($5) ~ tolower(q)) {print $1 \"---\" $2 \"---\" $3 \"---\" $4 \"---\" $5}' | head -n 50", psCmd, safeQuery)
	}
	out, err := session.RunCmd(cmd)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var processes []ProcessInfo
	for _, line := range lines {
		parts := strings.Split(line, "---")
		if len(parts) < 5 {
			continue
		}
		processes = append(processes, ProcessInfo{
			PID:     parts[0],
			User:    parts[1],
			CPU:     parts[2],
			Mem:     parts[3],
			Command: parts[4],
		})
	}

	return processes, nil
}
