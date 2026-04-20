# Data Model: Telemetry Hardening

## Backend Database Schema (SQLite)

### `agent_tokens` (New)
Manages the One-Time-Tokens (OTT) used for CSR authorization.
- `id` (TEXT, Primary Key, UUID)
- `token_hash` (TEXT, NOT NULL): SHA256 hash of the generated OTT.
- `machine_id` (TEXT, NOT NULL): The expected VM identifier.
- `created_at` (DATETIME, NOT NULL): Timestamp of creation.
- `expires_at` (DATETIME, NOT NULL): `created_at` + 10 minutes.
- `used_at` (DATETIME, NULL): Timestamp when the token was successfully used (null if unused).

### `agent_certificates` (New/Updated)
Tracks issued certificates and their revocation status.
- `serial_number` (TEXT, Primary Key): Certificate serial number.
- `machine_id` (TEXT, NOT NULL): The VM identifier this cert was issued to.
- `issued_at` (DATETIME, NOT NULL)
- `expires_at` (DATETIME, NOT NULL): `issued_at` + 30 days.
- `revoked_at` (DATETIME, NULL): Timestamp of revocation (adds the cert to the CRL).
- `revocation_reason` (TEXT, NULL)

## Agent Internal Models

### `WALEntry`
Represents a single batch written to disk.
- `Length` (uint32): Length of the payload.
- `Checksum` (uint32): CRC32 checksum of the payload.
- `Type` (uint8): 0x01 for Metrics, 0x02 for Logs.
- `Payload` ([]byte): Serialized protobuf message.

### `HealthMetrics`
Exposed via the agent's lightweight local HTTP endpoint.
- `ott_rejections` (int64)
- `crl_rejections` (int64)
- `wal_recoveries` (int64): Number of times the WAL was truncated/recovered on startup.
- `wal_evictions` (int64): Number of batches dropped due to the 50MB limit.
- `backpressure_activations` (int64): Number of times collection was throttled.
- `memory_usage_mb` (float64)
- `cpu_usage_percent` (float64)

## State Transitions

**OTT Lifecycle**:
`GENERATED` -> `USED` (if CSR successful within 10m) -> `EXPIRED` (if 10m elapses unused)

**WAL Lifecycle**:
`APPENDING` -> `FSYNC` (every 10th batch) -> `EVICTED` (if 50MB reached, drop oldest)
`STARTUP` -> `VERIFY_CHECKSUMS` -> `TRUNCATE_CORRUPTED` -> `READY`
