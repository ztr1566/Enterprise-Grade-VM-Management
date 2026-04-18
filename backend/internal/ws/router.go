package ws

import (
	"net/http"

	"github.com/gorilla/websocket"
)

// Upgrader handles the transition from HTTP to WebSocket protocols.
var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Note: In production, implement strict origin checking.
		return true
	},
}

// Upgrade upgrades an HTTP connection to a WebSocket connection.
func Upgrade(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	return Upgrader.Upgrade(w, r, nil)
}
