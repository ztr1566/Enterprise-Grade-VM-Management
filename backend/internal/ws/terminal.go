package ws

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	gossh "golang.org/x/crypto/ssh"
	"go.uber.org/zap"
	"backend/internal/audit"
	sshclient "backend/internal/ssh"
)

// TerminalHandler handles WebSocket connections that bridge to SSH sessions.
type TerminalHandler struct {
	DB *sql.DB
}

// wsMessage is the unified JSON protocol for all browser→backend messages.
//   - type "resize": cols/rows contain the terminal dimensions
//   - type "input":  data contains raw terminal keystrokes (string)
type wsMessage struct {
	Type string `json:"type"`
	Data string `json:"data,omitempty"`
	Cols int    `json:"cols,omitempty"`
	Rows int    `json:"rows,omitempty"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// TODO: restrict to known origins in production
		return true
	},
}

// ServeTerminal upgrades HTTP to WebSocket, performs a dimension handshake,
// allocates the PTY, starts the shell, and audits the full session lifecycle.
func (h *TerminalHandler) ServeTerminal(w http.ResponseWriter, r *http.Request) {
	vmID := r.PathValue("id")
	if vmID == "" {
		http.Error(w, "Missing VM ID", http.StatusBadRequest)
		return
	}

	// Extract authenticated user ID injected by AuthMiddleware
	userID, _ := r.Context().Value("user_id").(string)
	if userID == "" {
		userID = "unknown"
	}

	// 1. Upgrade HTTP → WebSocket
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		audit.Logger.Error("WebSocket upgrade failed",
			zap.String("vmID", vmID), zap.String("userID", userID), zap.Error(err))
		return
	}
	defer wsConn.Close()

	audit.Logger.Info("Terminal WebSocket connected",
		zap.String("vmID", vmID), zap.String("userID", userID))

	// 2. Establish SSH connection (fetches VM, decrypts credential, dials host)
	loginAs := r.URL.Query().Get("login_as")
	sshSession, err := sshclient.NewSession(h.DB, vmID, loginAs)
	if err != nil {
		// AUDIT: SSH_CONNECTION_FAILED — credential errors, network failures, etc.
		audit.LogSSHConnectionFailed(h.DB, userID, vmID, err.Error())
		wsConn.WriteMessage(websocket.TextMessage,
			[]byte("ERROR: SSH connection failed: "+err.Error()+"\r\n"))
		return
	}
	defer sshSession.Close()

	// 3. HANDSHAKE: Wait for the first message from the client — must be a resize event
	//    with the actual rendered terminal dimensions.
	wsConn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, raw, err := wsConn.ReadMessage()
	wsConn.SetReadDeadline(time.Time{}) // clear deadline
	if err != nil {
		audit.Logger.Error("Handshake timeout: no initial resize received",
			zap.String("vmID", vmID), zap.String("userID", userID), zap.Error(err))
		audit.LogSSHConnectionFailed(h.DB, userID, vmID, "handshake timeout: "+err.Error())
		return
	}

	var initMsg wsMessage
	cols, rows := 80, 24 // safe defaults
	if jsonErr := json.Unmarshal(raw, &initMsg); jsonErr == nil &&
		initMsg.Type == "resize" && initMsg.Cols > 0 && initMsg.Rows > 0 {
		cols = initMsg.Cols
		rows = initMsg.Rows
	}

	audit.Logger.Info("Handshake received initial dimensions",
		zap.String("vmID", vmID), zap.Int("cols", cols), zap.Int("rows", rows))

	// 4. Request a PTY with the client's actual terminal size
	modes := gossh.TerminalModes{gossh.ECHO: 1}
	if err := sshSession.RequestPty("xterm-256color", rows, cols, modes); err != nil {
		audit.LogSSHConnectionFailed(h.DB, userID, vmID, "PTY request failed: "+err.Error())
		wsConn.WriteMessage(websocket.TextMessage,
			[]byte("ERROR: Failed to allocate PTY: "+err.Error()+"\r\n"))
		return
	}

	// 5. Wire up SSH stdin/stdout pipes before starting the shell
	sshIn, err := sshSession.StdinPipe()
	if err != nil {
		audit.LogSSHConnectionFailed(h.DB, userID, vmID, "stdin pipe failed: "+err.Error())
		wsConn.WriteMessage(websocket.TextMessage, []byte("ERROR: Failed to open stdin pipe\r\n"))
		return
	}
	defer sshIn.Close()

	sshOut, err := sshSession.StdoutPipe()
	if err != nil {
		audit.LogSSHConnectionFailed(h.DB, userID, vmID, "stdout pipe failed: "+err.Error())
		wsConn.WriteMessage(websocket.TextMessage, []byte("ERROR: Failed to open stdout pipe\r\n"))
		return
	}

	// 6. Start the interactive login shell AFTER handshake and PTY are established
	if err := sshSession.Shell(); err != nil {
		audit.LogSSHConnectionFailed(h.DB, userID, vmID, "shell start failed: "+err.Error())
		wsConn.WriteMessage(websocket.TextMessage,
			[]byte("ERROR: Failed to start shell: "+err.Error()+"\r\n"))
		return
	}

	// AUDIT: SSH_SESSION_STARTED — shell is live, session is now active
	sessionStart := time.Now()
	audit.LogSSHSessionStarted(h.DB, userID, vmID)

	// GUARANTEE: SSH_SESSION_ENDED is logged even if a panic occurs in a goroutine.
	// defer runs as the function returns (normal, error, or recovered panic).
	defer func() {
		if r := recover(); r != nil {
			audit.Logger.Error("Recovered panic in terminal session",
				zap.String("vmID", vmID), zap.String("userID", userID),
				zap.Any("panic", r))
		}
		duration := time.Since(sessionStart)
		audit.LogSSHSessionEnded(h.DB, userID, vmID, duration)
		audit.Logger.Info("Terminal session ended",
			zap.String("vmID", vmID), zap.String("userID", userID),
			zap.Int64("duration_seconds", int64(duration.Seconds())))
	}()

	audit.Logger.Info("SSH shell started — bridging WebSocket ↔ SSH",
		zap.String("vmID", vmID), zap.String("userID", userID),
		zap.Int("cols", cols), zap.Int("rows", rows))

	done := make(chan struct{}, 2)

	// 7a. SSH stdout → WebSocket (raw binary — xterm renders this directly)
	go func() {
		defer func() { done <- struct{}{} }()
		buf := make([]byte, 32*1024)
		for {
			n, err := sshOut.Read(buf)
			if n > 0 {
				if writeErr := wsConn.WriteMessage(websocket.BinaryMessage, buf[:n]); writeErr != nil {
					return
				}
			}
			if err != nil {
				if err != io.EOF {
					audit.Logger.Warn("SSH stdout read closed",
						zap.String("vmID", vmID), zap.Error(err))
				}
				return
			}
		}
	}()

	// 7b. WebSocket → SSH stdin or window resize
	go func() {
		defer func() { done <- struct{}{} }()
		for {
			_, raw, err := wsConn.ReadMessage()
			if err != nil {
				return
			}

			var msg wsMessage
			if jsonErr := json.Unmarshal(raw, &msg); jsonErr != nil {
				// Non-JSON fallback: treat as raw input
				if _, writeErr := sshIn.Write(raw); writeErr != nil {
					return
				}
				continue
			}

			switch msg.Type {
			case "input":
				if _, writeErr := sshIn.Write([]byte(msg.Data)); writeErr != nil {
					return
				}
			case "resize":
				if msg.Rows > 0 && msg.Cols > 0 {
					if resizeErr := sshSession.WindowChange(msg.Rows, msg.Cols); resizeErr != nil {
						audit.Logger.Warn("WindowChange failed",
							zap.String("vmID", vmID), zap.Error(resizeErr))
					} else {
						audit.Logger.Debug("Terminal resized",
							zap.String("vmID", vmID),
							zap.Int("cols", msg.Cols), zap.Int("rows", msg.Rows))
					}
				}
			default:
				audit.Logger.Warn("Unknown WS message type", zap.String("type", msg.Type))
			}
		}
	}()

	// Block until either side disconnects, then the deferred audit fires
	<-done
}
