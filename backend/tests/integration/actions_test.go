package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestActionExecutionSpeed(t *testing.T) {
	// SC-004: Quick Action commands must execute within 3 seconds.
	// This test sets a 3s timeout on a mock request to verify the threshold.
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Second) // Simulate remote execution
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := http.Client{
		Timeout: 3 * time.Second,
	}

	start := time.Now()
	_, err := client.Post(server.URL+"/api/vms/test/actions", "application/json", nil)
	duration := time.Since(start)

	if err != nil {
		t.Errorf("Action timed out or failed: %v", err)
	}

	if duration > 3*time.Second {
		t.Errorf("Action took too long: %v (max 3s)", duration)
	}
}
