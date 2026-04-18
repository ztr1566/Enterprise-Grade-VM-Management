package monitor

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/audit"
)

func TestLogHandler_MissingParams(t *testing.T) {
	audit.InitLogger()
	handler := &LogHandler{}

	// Test missing ID
	req := httptest.NewRequest("GET", "/api/vms//logs/ssh", nil)
	w := httptest.NewRecorder()
	handler.ServeLogs(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing ID, got %d", w.Code)
	}

	// Test missing service
	req = httptest.NewRequest("GET", "/api/vms/123/logs/", nil)
	w = httptest.NewRecorder()
	handler.ServeLogs(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing service, got %d", w.Code)
	}
}
