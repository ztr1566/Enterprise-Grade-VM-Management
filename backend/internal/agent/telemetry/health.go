package telemetry

import (
	"encoding/json"
	"net/http"

	"backend/internal/agent/governance"
)

// HealthHandler provides an HTTP endpoint for agent resource status.
type HealthHandler struct {
	Monitor *governance.Monitor
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cpu, ram, throttled := h.Monitor.GetStatus()

	status := "healthy"
	if throttled {
		status = "throttled"
	}

	response := map[string]interface{}{
		"status":    status,
		"cpu_usage": cpu,
		"ram_usage": ram,
		"limits": map[string]interface{}{
			"cpu": governance.MaxCPUPercent,
			"ram": governance.MaxRAMBytes,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if throttled {
		w.WriteHeader(http.StatusServiceUnavailable) // 503
	} else {
		w.WriteHeader(http.StatusOK)
	}
	json.NewEncoder(w).Encode(response)
}

// StartHealthServer launches a local HTTP server for health checks.
func StartHealthServer(addr string, monitor *governance.Monitor) {
	mux := http.NewServeMux()
	mux.Handle("/health", &HealthHandler{Monitor: monitor})
	
	go func() {
		if err := http.ListenAndServe(addr, mux); err != nil {
			// Local server failure isn't fatal but should be logged
		}
	}()
}
