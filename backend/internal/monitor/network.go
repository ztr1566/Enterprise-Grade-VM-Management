package monitor

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	sshclient "backend/internal/ssh"
)

type netSample struct {
	rxBytes   uint64
	txBytes   uint64
	timestamp time.Time
}

var (
	netHistory   = make(map[string]netSample)
	netHistoryMu sync.Mutex
)

// NetworkMetrics contains the calculated rates and connection counts.
type NetworkMetrics struct {
	RxBytesSec        float64 `json:"rx_bytes_sec"`
	TxBytesSec        float64 `json:"tx_bytes_sec"`
	ActiveConnections int     `json:"active_connections"`
}

// GetNetworkMetrics collects network data from the VM and calculates rates.
func GetNetworkMetrics(db *sql.DB, vmID string) (*NetworkMetrics, error) {
	session, err := sshclient.NewSession(db, vmID)
	if err != nil {
		return nil, fmt.Errorf("ssh connection failed: %w", err)
	}
	defer session.Close()

	// Command to get rx/tx bytes (sum of all physical interfaces) and active TCP connections
	// Filtering: lo, docker, veth, br-, virbr
	cmd := `cat /proc/net/dev | grep -vE 'lo|docker|veth|br-|virbr' | awk 'NR>2 {rx+=$2; tx+=$10} END {print rx " " tx}' && ss -Htn state established | wc -l`
	out, err := session.RunCmd(cmd)
	if err != nil {
		return nil, fmt.Errorf("command execution failed: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("unexpected output format: %s", string(out))
	}

	// Parse bytes
	byteParts := strings.Fields(lines[0])
	if len(byteParts) < 2 {
		return nil, fmt.Errorf("invalid byte stats: %s", lines[0])
	}
	rx, _ := strconv.ParseUint(byteParts[0], 10, 64)
	tx, _ := strconv.ParseUint(byteParts[1], 10, 64)

	// Parse connections
	conns, _ := strconv.Atoi(strings.TrimSpace(lines[1]))
	// ss output includes header, so if wc -l is used on established, it might be off by 1 if there are results
	// Actually ss -tun state established | wc -l returns 0 if none, or 1+ (count)
	// If there are established connections, the output of ss includes a header line.
	// Let's refine the command: ss -Htn state established | wc -l (-H suppresses header)
	
	now := time.Now()
	metrics := &NetworkMetrics{
		ActiveConnections: conns,
	}

	netHistoryMu.Lock()
	prev, exists := netHistory[vmID]
	netHistory[vmID] = netSample{rxBytes: rx, txBytes: tx, timestamp: now}
	netHistoryMu.Unlock()

	if exists {
		duration := now.Sub(prev.timestamp).Seconds()
		if duration > 0 {
			metrics.RxBytesSec = float64(rx-prev.rxBytes) / duration
			metrics.TxBytesSec = float64(tx-prev.txBytes) / duration
		}
	}

	// Sanity check: if counter wrapped or reset, set to 0
	if metrics.RxBytesSec < 0 {
		metrics.RxBytesSec = 0
	}
	if metrics.TxBytesSec < 0 {
		metrics.TxBytesSec = 0
	}

	return metrics, nil
}
