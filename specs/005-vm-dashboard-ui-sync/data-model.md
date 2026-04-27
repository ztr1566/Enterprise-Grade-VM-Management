# Data Model: VM Dashboard UI Redesign & Backend Sync

**Branch**: `005-vm-dashboard-ui-sync`  
**Date**: 2026-04-27  

---

## Frontend TypeScript Interfaces

These interfaces define the data shapes consumed by the frontend from the backend API. No backend changes are made; these are documentation of existing contracts.

### VM (Inventory List Item)

```typescript
interface VM {
  id: string;
  name: string;
  host: string;
  management_username: string;
  auth_type: string;           // "password" | "key"
  status: string;              // "online" | "offline"
  provisioning_status: string; // "provisioned" | "pending" | "not_provisioned" | "failed"
  tags: string[];
  authorized_users: string[];
  key_id?: string;
  created_at: string;          // ISO 8601
}
```

**Source**: `GET /api/vms` returns `VM[]`  
**Source**: `GET /api/vms/{id}` returns `VM` (includes `credential` field, excluded from list)

### VMStats (Per-Card Telemetry)

```typescript
interface VMStats {
  cpu: number;                // 0-100 percentage
  ram: number;                // 0-100 percentage
  disk: number;               // 0-100 percentage
  network: {
    rx_bytes_sec: number;
    tx_bytes_sec: number;
    active_connections: number;
  };
  processes: ProcessInfo[];
  timestamp: string;          // ISO 8601
}

interface ProcessInfo {
  pid: string;
  user: string;
  cpu: string;
  mem: string;
  command: string;
}
```

**Source**: `GET /api/vms/{id}/stats`  
**Usage**: VM card (cpu, ram only) + VM detail page (all fields)

### DashboardStats (Aggregate Overview)

```typescript
interface DashboardStats {
  total_vms: number;
  online_vms: number;
  offline_vms: number;
  provisioned: number;
  pending: number;
  failed: number;
  audit_events: number;
}
```

**Source**: `GET /api/dashboard/stats`

### AuditLog

```typescript
interface AuditLog {
  id: number;
  event_type: string;
  details: string;
  timestamp: string;
}
```

**Source**: `GET /api/audit-logs?limit={n}&offset={n}` returns `{ logs: AuditLog[], total: number }`

### CreateVMRequest / UpdateVMRequest

```typescript
interface CreateVMRequest {
  name: string;
  host: string;
  management_username: string;
  auth_type: string;
  credential: string;
  key_id?: string;
  tags: string[];
}

interface UpdateVMRequest {
  name: string;
  host: string;
  management_username: string;
  auth_type: string;
  credential?: string;    // If empty, credential unchanged
  key_id?: string;
  tags: string[];
}
```

**Target**: `POST /api/vms` (create) and `PUT /api/vms/{id}` (update)

### PowerRequest

```typescript
interface PowerRequest {
  action: string;  // "reboot" | "shutdown" | "sleep"
}
```

**Target**: `POST /api/vms/{id}/power`

---

## Status State Machines

### VM Connectivity Status

```
offline ──(agent heartbeat received)──> online
online ──(heartbeat timeout)──> offline
```

- Set by backend based on agent telemetry heartbeat
- Frontend reads `vm.status` — no write capability

### VM Provisioning Status

```
not_provisioned ──(provision triggered)──> pending
pending ──(provisioning completes)──> provisioned
pending ──(provisioning fails)──> failed
failed ──(retry triggered)──> pending
```

- Transitions triggered by `POST /api/vms/{id}/provision`
- Frontend reads `vm.provisioning_status` and provides retry action for `failed` state

---

## Component Data Dependencies

| Component | Primary Data | Secondary Data | Refresh Strategy |
|-----------|-------------|---------------|-----------------|
| DashboardView | DashboardStats | — | 15s polling |
| Inventory (grid) | VM[] | VMStats (per online VM) | 5s/10s polling (list) + staggered stats |
| VM Card | VM | VMStats (optional) | From parent poll |
| VMDetailView | VM | VMStats, Processes, Services, Logs | Per-tab auto-refresh |
| AuditLogPage | AuditLog[] | — | On-demand pagination |
| KeysManagement | SSHKey[] | — | On-demand |
| AddVmModal | — (form) | SSHKey[] (for key_id) | On open |
| Sidebar | — | — | Static |
| Layout/TopBar | DashboardStats (server count) | — | 15s polling |
