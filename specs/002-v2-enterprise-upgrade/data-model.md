# Data Model: V2 Enterprise Upgrade

**Feature**: V2 Enterprise Upgrade  
**Date**: 2026-04-18  
**Spec**: [spec.md](/home/ztr/Projects/new_project/specs/002-v2-enterprise-upgrade/spec.md)

## Entity Definitions

### Provisioning Record (New)

Tracks the provisioning state of each VM.

| Field              | Type      | Constraints                          | Description                                    |
|--------------------|-----------|--------------------------------------|------------------------------------------------|
| vm_id              | TEXT      | PK, FK → vms(id), ON DELETE CASCADE | Foreign key to the VM this record belongs to   |
| status             | TEXT      | NOT NULL, CHECK IN ('pending','provisioned','failed','not_provisioned') | Current provisioning state |
| last_attempt_at    | DATETIME  | NULLABLE                             | Timestamp of the most recent provisioning attempt |
| error_message      | TEXT      | NULLABLE                             | Error details from the last failed attempt     |
| provisioned_user   | TEXT      | NULLABLE                             | The SSH username the sudoers rules were created for |

**State Transitions**:
```
not_provisioned → pending → provisioned
                        ↘ failed → pending (retry)
```

**Migration**: `003_provisioning.sql`

```sql
CREATE TABLE IF NOT EXISTS provisioning_records (
    vm_id           TEXT PRIMARY KEY REFERENCES vms(id) ON DELETE CASCADE,
    status          TEXT NOT NULL DEFAULT 'not_provisioned',
    last_attempt_at DATETIME,
    error_message   TEXT,
    provisioned_user TEXT
);
```

---

### VM (Existing — Extended)

The existing `vms` table remains unchanged. The `provisioning_records` table references it via `vm_id` FK with cascade delete, maintaining clean separation.

No schema changes to the `vms` table are required. The provisioning status is joined at query time.

---

### Network Metric (New — In-memory only)

Network metrics are collected in real-time and are NOT persisted to the database. They exist only in the backend's memory during a monitoring session and are pushed to the frontend via the existing stats polling mechanism.

| Field         | Type    | Description                            |
|---------------|---------|----------------------------------------|
| vm_id         | TEXT    | The VM this metric belongs to          |
| rx_bytes_sec  | FLOAT   | Received bytes per second              |
| tx_bytes_sec  | FLOAT   | Transmitted bytes per second           |
| active_conns  | INT     | Count of established TCP connections   |
| timestamp     | TIME    | When this sample was taken             |

---

### Process Snapshot (New — In-memory only)

Process snapshots are collected on-demand and are NOT persisted.

| Field     | Type    | Description                          |
|-----------|---------|--------------------------------------|
| pid       | INT     | Process ID                           |
| user      | TEXT    | Owner of the process                 |
| cpu_pct   | FLOAT   | CPU usage percentage                 |
| mem_pct   | FLOAT   | Memory usage percentage              |
| command   | TEXT    | Command name / path                  |

---

## Relationships

```
vms (1) ──── (0..1) provisioning_records
```

- One VM has at most one provisioning record.
- Deleting a VM cascades to delete its provisioning record.
- Network metrics and process snapshots are transient — no DB persistence, no relationships.
