# Phase 1: Data Model

## Entities

### VM Agent Identity
- **Private Key**: Locally generated on the VM (e.g., ECDSA or Ed25519). Never leaves the VM.
- **CSR**: Certificate Signing Request containing the VM identifier.
- **mTLS Certificate**: Issued by the backend CA, used for all subsequent gRPC connections.

### Telemetry Batch Payload
- **Metrics Batch**: Array of hardware metric samples.
  - CPU usage percentage
  - RAM usage (bytes/percentage)
  - Disk usage/IO
  - Network I/O (bytes tx/rx)
  - Timestamp
- **Logs Batch**: Array of structured logs.
  - Timestamp
  - Severity (INFO, WARN, ERROR, etc.)
  - Component (e.g., system, application name)
  - Message (JSON or text)

### Write-Ahead Log (WAL) Entry
- **Index**: Sequential ID.
- **Payload**: Serialized batched telemetry (Metrics or Logs).
- **Status**: Pending or Acknowledged.