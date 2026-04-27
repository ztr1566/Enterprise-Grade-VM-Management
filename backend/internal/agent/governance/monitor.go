package governance

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

const (
	MaxRAMBytes = 100 * 1024 * 1024 // 100MB
	MaxCPUPercent = 10.0           // 10%
)

// Monitor tracks the agent's own resource consumption.
type Monitor struct {
	mu           sync.RWMutex
	backpressure bool
	cpuUsage     float64
	ramUsage     uint64
}

// NewMonitor creates a new resource monitor.
func NewMonitor() *Monitor {
	return &Monitor{}
}

// Start launches the monitoring loop.
func (m *Monitor) Start(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	p, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		log.Printf("Governance: Failed to get process info: %v", err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.checkResources(p)
		}
	}
}

func (m *Monitor) checkResources(p *process.Process) {
	cpu, err := p.Percent(0)
	if err != nil {
		log.Printf("Governance: Failed to measure CPU: %v", err)
	}

	mem, err := p.MemoryInfo()
	if err != nil {
		log.Printf("Governance: Failed to measure RAM: %v", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.cpuUsage = cpu
	if mem != nil {
		m.ramUsage = mem.RSS
	}

	// Apply backpressure if ANY limit is exceeded
	m.backpressure = m.cpuUsage > MaxCPUPercent || m.ramUsage > MaxRAMBytes

	if m.backpressure {
		log.Printf("Governance: [BACKPRESSURE ACTIVE] CPU: %.2f%%, RAM: %dMB", m.cpuUsage, m.ramUsage/(1024*1024))
	}
}

// IsThrottled returns true if the agent is under resource backpressure.
func (m *Monitor) IsThrottled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.backpressure
}

// GetStatus returns the current resource metrics.
func (m *Monitor) GetStatus() (float64, uint64, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cpuUsage, m.ramUsage, m.backpressure
}
