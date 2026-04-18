# Technical Research & Decisions (Revision 2)

## Technology Stack Choices

**Decision:** Go (Golang) backend, React + Tailwind CSS frontend
**Rationale:** Go provides industry-leading native concurrency and the highly robust `golang.org/x/crypto/ssh` standard library out of the box, making it exceptionally well-suited for safely multiplexing dozens of persistent SSH PTY streams and WebSocket connections simultaneously with minimal resource overhead. React satisfies the frontend component constraints, and Tailwind fulfills the aesthetic directive natively.
**Alternatives considered:** Node.js with `ssh2`. Discarded via explicit user constraint in favor of Go's native capabilities.

## Communication Layer

**Decision:** `gorilla/websocket` for Go; `xterm.js` for Frontend
**Rationale:** The `gorilla/websocket` library is the most established, battle-tested standard for handling WebSockets in Go. HTTP polling is strictly prohibited. For the terminal interface, `xterm.js` is the industry standard and natively supports `xterm-addon-attach` to read unadulterated `[]byte` streams dispatched directly via WebSocket from the Go `crypto/ssh` sessions.

## Credential Security & Storage

**Decision:** SQLite with AES-256-GCM symmetric field-level encryption via Go's `crypto/aes` and `crypto/cipher` packages
**Rationale:** The platform requires a lightweight embedded DB (SQLite) to store host metadata, but directly forbids plaintext passwords. Passwords and SSH Private Keys will be encrypted in Go memory before writing to the database, ensuring that an attacker gaining a copy of the `.sqlite` file cannot compromise the remote infrastructure.

## Telemetry Collection

**Decision:** Agentless collection via multiplexed SSH execution
**Rationale:** `golang.org/x/crypto/ssh` easily supports opening multiple concurrent `ssh.Session` channels on a single authenticated `ssh.Client` connection. One channel serves the `xterm` PTY, while separate lightweight channels periodically execute standard Linux commands (`top -bn1`, `cat /proc/stat`) to stream parsed metrics through JSON WebSockets directly to the React dashboard.

## Audit Logging

**Decision:** `uber-go/zap` for structured JSON audit logging
**Rationale:** The constitution mandates an immutable audit trail for credential access. `uber-go/zap` provides extremely fast, memory-efficient structured logging that is easy to pipe into secondary log aggregators or SIEMs without dragging down API performance.
