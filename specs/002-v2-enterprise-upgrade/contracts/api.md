# API Contracts: V2 Enterprise Upgrade

**Date**: 2026-04-18

## New Endpoints

### POST /api/vms (Modified — Provisioning Integration)

The existing `POST /api/vms` endpoint is extended. After persisting the VM record, the backend asynchronously triggers the provisioning engine. The response now includes a `provisioning_status` field.

**Response** (201 Created):
```json
{
  "id": "uuid",
  "name": "string",
  "host": "string",
  "username": "string",
  "status": "offline",
  "provisioning_status": "pending",
  "tags": ["string"]
}
```

---

### POST /api/vms/{id}/provision (New)

Manually triggers re-provisioning for a VM that previously failed or needs refreshing.

**Request**: No body required.

**Response** (200 OK):
```json
{
  "vm_id": "uuid",
  "provisioning_status": "pending",
  "message": "Provisioning initiated"
}
```

**Error** (404):
```json
{
  "error": "VM not found"
}
```

---

### GET /api/vms/{id}/stats (Modified — Extended Metrics)

The existing stats endpoint is extended to include network and process data.

**Response** (200 OK):
```json
{
  "cpu": 45.2,
  "ram": 68.1,
  "disk": 32.0,
  "network": {
    "rx_bytes_sec": 125000.5,
    "tx_bytes_sec": 42000.3,
    "active_connections": 47
  },
  "processes": [
    {
      "pid": 1234,
      "user": "root",
      "cpu_pct": 12.5,
      "mem_pct": 3.2,
      "command": "/usr/bin/nginx"
    }
  ],
  "timestamp": "2026-04-18T00:00:00Z"
}
```

---

### GET /api/vms/{id}/provisioning (New)

Returns the current provisioning status for a specific VM.

**Response** (200 OK):
```json
{
  "vm_id": "uuid",
  "status": "provisioned",
  "last_attempt_at": "2026-04-18T00:00:00Z",
  "error_message": null,
  "provisioned_user": "admin"
}
```

---

### GET /api/audit (New)

Returns the full audit log with pagination.

**Query Parameters**:
- `page` (int, default: 1)
- `limit` (int, default: 50)

**Response** (200 OK):
```json
{
  "entries": [
    {
      "id": 1,
      "event_type": "VM_PROVISIONED",
      "details": "Provisioning completed for VM abc-123",
      "timestamp": "2026-04-18T00:00:00Z"
    }
  ],
  "total": 100,
  "page": 1,
  "limit": 50
}
```

## New Audit Event Types

| Event Type              | Trigger                                |
|-------------------------|----------------------------------------|
| `VM_PROVISIONING_START` | Provisioning engine begins execution   |
| `VM_PROVISIONED`        | Provisioning completes successfully    |
| `VM_PROVISIONING_FAILED`| Provisioning fails (with error detail) |
| `VM_REPROVISION_REQUESTED` | Manual re-provision triggered       |

## Security Constraints (Enforced)

- ❌ No `POST /api/vms/{id}/sudo` endpoint exists or will be created.
- ❌ No API endpoint accepts a `sudo_password` field.
- ❌ No WebSocket message carries sudo credentials.
- ✅ All sudo operations are pre-authorized on the target VM via provisioning.
