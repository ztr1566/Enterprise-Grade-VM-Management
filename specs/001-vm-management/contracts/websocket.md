# WebSocket Event Contracts

The platform uses a WebSocket connection between the frontend and the Node.js backend to facilitate all Real-time Telemetry and SSH Terminal functionality.

## Namespace Structure
The WebSocket router should split functionality into two namespaces (or distinct message types over a single socket):
1. `Telemetry` (`/ws/telemetry`)
2. `Terminal` (`/ws/terminal`)

## Telemetry Events

### Server -> Client Broadcasts

**`telemetry.update`**
Emitted every 5 seconds per connected VM to send active metrics to the dashboard.
```json
{
  "event": "telemetry.update",
  "payload": {
    "vmId": "uuid-string",
    "cpuPercent": 12.5,
    "ramPercent": 45.2,
    "networkTx": 1024,
    "networkRx": 2048,
    "status": "ONLINE"
  }
}
```

**`service.health`**
Emitted when a tracked daemon changes state on a monitored VM.
```json
{
  "event": "service.health",
  "payload": {
    "vmId": "uuid-string",
    "serviceName": "nginx.service",
    "status": "Running"
  }
}
```

## Terminal Events

### Client -> Server
**`terminal.input`**
Emitted when the user types in the `xterm.js` window.
```json
{
  "event": "terminal.input",
  "payload": {
    "vmId": "uuid-string",
    "data": "cd /var/log\n" // Raw character stroke or string
  }
}
```

**`terminal.resize`**
Emitted when the browser window resizes, modifying the PTY size on the remote host.
```json
{
  "event": "terminal.resize",
  "payload": {
    "vmId": "uuid-string",
    "cols": 80,
    "rows": 24
  }
}
```

### Server -> Client
**`terminal.output`**
Emitted when the SSH process flushes stdout/stderr buffers back to the terminal.
```json
{
  "event": "terminal.output",
  "payload": {
    "vmId": "uuid-string",
    "data": "\\x1b[32muser@host:~$\\x1b[0m "
  }
}
```
