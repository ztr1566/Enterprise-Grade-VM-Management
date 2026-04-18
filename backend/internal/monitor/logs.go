package monitor

import (
	"database/sql"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"backend/internal/audit"
	sshclient "backend/internal/ssh"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// TODO: restrict to known origins in production
		return true
	},
}

// LogHandler handles WebSocket connections that stream systemd logs.
type LogHandler struct {
	DB *sql.DB
}

// ServeLogs upgrades HTTP to WebSocket and streams journalctl output via SSH.
func (h *LogHandler) ServeLogs(w http.ResponseWriter, r *http.Request) {
	vmID := r.PathValue("id")
	service := r.PathValue("service")
	if vmID == "" || service == "" {
		http.Error(w, "Missing VM ID or service name", http.StatusBadRequest)
		return
	}

	// 1. Upgrade to WebSocket
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		audit.Logger.Error("Log WebSocket upgrade failed", zap.Error(err))
		return
	}
	defer wsConn.Close()

	// 2. Establish SSH connection
	sshSession, err := sshclient.NewSession(h.DB, vmID)
	if err != nil {
		wsConn.WriteMessage(websocket.TextMessage, []byte("ERROR: SSH connection failed: "+err.Error()+"\r\n"))
		return
	}
	defer sshSession.Close()

	// 3. Setup pipe
	stdout, err := sshSession.StdoutPipe()
	if err != nil {
		wsConn.WriteMessage(websocket.TextMessage, []byte("ERROR: Failed to open stdout pipe\r\n"))
		return
	}

	// 4. Start journalctl -f
	// We use -n 100 to show some history, and -f to follow.
	cmd := "sudo journalctl -u " + service + " -f -n 100 --no-pager"
	if err := sshSession.Start(cmd); err != nil {
		wsConn.WriteMessage(websocket.TextMessage, []byte("ERROR: Failed to start journalctl: "+err.Error()+"\r\n"))
		return
	}

	// 5. Stream from SSH stdout to WebSocket
	done := make(chan struct{})
	go func() {
		defer close(done)
		buf := make([]byte, 16*1024)
		for {
			n, err := stdout.Read(buf)
			if n > 0 {
				if writeErr := wsConn.WriteMessage(websocket.TextMessage, buf[:n]); writeErr != nil {
					return
				}
			}
			if err != nil {
				if err != io.EOF {
					audit.Logger.Warn("Log stream read error", zap.Error(err))
				}
				return
			}
		}
	}()

	// 6. Keep-alive and wait for disconnect
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			if err := wsConn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
