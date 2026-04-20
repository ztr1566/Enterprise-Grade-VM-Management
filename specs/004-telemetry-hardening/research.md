# Phase 0: Research & Architecture Decisions

## 1. Resource Governance (CPU/RAM Limits)

**Decision**: Use an internal ticker-based resource monitor utilizing `github.com/shirou/gopsutil/v3` and `runtime` packages, rather than cgroups.

**Rationale**: 
The agent needs to run on standard Linux and Containerized Workloads. Relying on host-level cgroups requires elevated permissions and complex OS-specific configuration during provisioning. An internal Go ticker that periodically checks `runtime.MemStats` and `gopsutil.Process` CPU usage is cross-platform, self-contained, and perfectly fits a daemon architecture. 
When limits (100MB RAM, 10% CPU) are approached, the agent will dynamically adjust its telemetry collection intervals (backpressure) or reduce batch sizes before eviction policies kick in.

**Alternatives considered**: 
- *Systemd cgroups*: Rejected because it tightly couples the agent to systemd and requires root/sudo access to configure limits during provisioning, violating the "simple binary SCP" requirement.
- *Docker resource limits*: Rejected as the agent must also run natively on non-containerized standard Linux VMs.

## 2. WAL Integrity & Recovery

**Decision**: Append CRC32 checksums to each WAL batch entry, use `os.File.Sync()` every 10 batches, and implement a truncating recovery scan on startup.

**Rationale**:
Using `hash/crc32` provides fast, reliable corruption detection. By fsyncing every 10th batch, we strike an optimal balance between I/O performance and data durability. During startup, the WAL reader will verify checksums sequentially. If a checksum fails (e.g., partial write during power loss), the file is truncated at the last valid boundary. If the header is unreadable, the file is archived and recreated.

**Alternatives considered**:
- *SHA256*: Rejected as computationally too expensive for an agent with a 10% CPU limit, and cryptographic collision resistance is not required for local file integrity against crash corruption.
- *Fsync every batch*: Rejected as it would cause excessive disk I/O, impacting the host VM's disk performance.

## 3. Security: One-Time-Token (OTT) & Revocation

**Decision**: Store OTTs and the Certificate Revocation List (CRL) in the backend SQLite database. Use standard ECDSA P-256 for key generation on the agent.

**Rationale**:
SQLite is already the backend storage. Creating an `agent_tokens` table with an `expires_at` (10 minutes) and `used` boolean allows atomic validation of OTTs. Similarly, an `agent_crl` table allows immediate revocation. On every gRPC connection, a unary interceptor will check the client certificate's serial number against the cached CRL.

**Alternatives considered**:
- *In-memory map for OTTs*: Rejected because it would not survive a backend server restart, potentially locking out legitimately provisioned agents.
- *Redis for CRL/OTT*: Rejected as it introduces an unnecessary external dependency, violating the existing SQLite-centric architecture.

## 4. Protocol: gRPC Payload Enforcement

**Decision**: Implement a gRPC interceptor on the backend that rejects payloads > 4MB, and add client-side logic to split batches before transmission.

**Rationale**:
The default gRPC max message size is 4MB. Instead of increasing it (which risks memory exhaustion), we enforce it explicitly. The agent will measure the proto payload size before sending. If it exceeds 4MB, the slice of events is split into multiple smaller `StreamMetrics` or `StreamLogs` calls.
