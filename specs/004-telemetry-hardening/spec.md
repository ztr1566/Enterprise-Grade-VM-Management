# Feature Specification: Telemetry Agent Hardening

**Feature Branch**: `004-telemetry-hardening`  
**Created**: April 20, 2026  
**Status**: Draft  
**Input**: User description: "Create a hardening specification (004) that resolves all technical gaps and failed validation criteria identified in Phase 1. Security Hardening, Infrastructure Resilience, Protocol Rigor, and Network Fault Tolerance."

## Clarifications

### Session 2026-04-20

- Q: What is the OTT expiration window (how long a token remains valid if unused)? → A: 10-minute expiration window.
- Q: What observability signals should the agent/backend emit for hardening events? → A: Structured log events + lightweight health endpoint with key counters (OTT rejections, CRL hits, WAL recoveries, backpressure activations).
- Q: How is the renewal CSR authorized when the original OTT is consumed? → A: Agent authenticates renewal CSR using its current (still-valid) mTLS certificate — no new OTT needed.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - mTLS Certificate Lifecycle Management (Priority: P1)

As a platform operator, I want every agent certificate to expire automatically after 30 days and be renewable through a secure re-enrollment flow, so that compromised or stale credentials cannot persist indefinitely in the fleet.

**Why this priority**: Without a finite certificate lifetime, a compromised agent identity could remain trusted forever. This is the highest-impact security gap from Phase 1.

**Independent Test**: Can be independently tested by provisioning an agent, verifying its certificate has a 30-day expiry, waiting for (or simulating) expiry, and confirming that the agent automatically re-enrolls for a fresh certificate without manual intervention.

**Acceptance Scenarios**:

1. **Given** a newly provisioned agent completes the CSR flow, **When** the backend signs the certificate, **Then** the issued certificate has a validity period of exactly 30 days from issuance.
2. **Given** an agent whose certificate is within 24 hours of expiration, **When** the agent detects the approaching expiry, **Then** it authenticates using its current mTLS certificate to submit a renewal CSR and obtains a new 30-day certificate without telemetry interruption.
3. **Given** an agent presenting an expired certificate, **When** it attempts to stream telemetry, **Then** the backend rejects the connection and the agent falls back to its re-enrollment flow.

---

### User Story 2 - One-Time-Token CSR Authorization (Priority: P1)

As a security engineer, I want each CSR to be authorized by a single-use, time-limited token generated during provisioning, so that only legitimately deployed agents can obtain certificates and rogue CSR submissions are rejected.

**Why this priority**: The Phase 1 CSR flow lacked an authorization gate, meaning any entity that could reach the CSR endpoint could request a certificate. This is a critical zero-trust gap.

**Independent Test**: Can be tested by provisioning an agent with an OTT, verifying the CSR succeeds with a valid token, and then attempting a second CSR with the same token to confirm rejection.

**Acceptance Scenarios**:

1. **Given** the backend generates a One-Time-Token during agent provisioning, **When** the agent submits its CSR with this OTT, **Then** the backend validates the token, signs the certificate, and immediately invalidates the token.
2. **Given** a used or expired OTT, **When** any entity submits a CSR with it, **Then** the backend rejects the request and logs a security event.
3. **Given** no OTT is provided with a CSR, **When** the request reaches the backend, **Then** it is rejected with an authorization error.

---

### User Story 3 - Agent Resource Governance (Priority: P1)

As a VM owner, I want the telemetry agent to operate within strict resource limits (100MB RAM and 10% CPU), so that the agent never degrades the performance of my production workloads.

**Why this priority**: An unbounded agent could starve production applications of resources. This is a critical operational gap identified in Phase 1 validation.

**Independent Test**: Can be tested by running the agent under sustained high-telemetry load and monitoring that it self-throttles to stay within the defined resource envelope.

**Acceptance Scenarios**:

1. **Given** the agent is running under normal telemetry load, **When** resource consumption is measured, **Then** memory usage stays below 100MB and CPU usage stays below 10% of available cores.
2. **Given** the agent approaches its 100MB memory limit, **When** it detects the threshold, **Then** it activates backpressure (reducing batch sizes or collection frequency) rather than exceeding the limit.
3. **Given** a sustained CPU spike from telemetry processing, **When** CPU consumption approaches 10%, **Then** the agent throttles its collection intervals to remain within budget.

---

### User Story 4 - WAL Resilience and Recovery (Priority: P2)

As an operations engineer, I want the Write-Ahead Log to guarantee data durability through periodic fsync and to recover gracefully from corruption, so that telemetry data survives crashes and the agent self-heals without manual intervention.

**Why this priority**: Phase 1 defined a WAL but lacked fsync semantics, eviction policy specifics, and corruption recovery—all required for production-grade data integrity.

**Independent Test**: Can be tested by killing the agent process mid-write, restarting it, and verifying that data up to the last synced batch is recovered and that any corrupted trailing entries are safely truncated.

**Acceptance Scenarios**:

1. **Given** the agent is writing telemetry batches to the WAL, **When** every 10th batch is committed, **Then** the WAL performs an fsync to disk, ensuring durability of all preceding batches.
2. **Given** the WAL has reached its 50MB capacity, **When** new telemetry arrives, **Then** the oldest entries are evicted first ("drop-oldest" policy) to make room.
3. **Given** a WAL file with corrupted trailing entries (e.g., from a crash during write), **When** the agent starts up, **Then** it detects the corruption, truncates the damaged entries, recovers all valid preceding data, and resumes normal operation.
4. **Given** a WAL file that is entirely unreadable, **When** the agent starts up, **Then** it logs a critical error, archives the corrupted file for forensics, creates a fresh WAL, and continues operating.

---

### User Story 5 - Certificate Revocation (Priority: P2)

As a security administrator, I want to revoke a compromised agent's certificate immediately via a Certificate Revocation List, so that a stolen or decommissioned agent can no longer authenticate to the backend.

**Why this priority**: Revocation completes the certificate lifecycle. Without it, there is no mechanism to remove trust from a specific agent before its certificate expires.

**Independent Test**: Can be tested by revoking a running agent's certificate on the backend, then observing that the agent's next connection attempt is rejected.

**Acceptance Scenarios**:

1. **Given** an administrator revokes an agent's certificate, **When** the backend updates the CRL, **Then** the revoked certificate serial number appears on the CRL immediately.
2. **Given** an agent whose certificate serial is on the CRL, **When** it attempts to establish a connection, **Then** the backend rejects the handshake and logs a revocation event.
3. **Given** the CRL has been updated, **When** the backend checks a connecting agent's certificate, **Then** it validates the certificate against the current CRL on every connection attempt.

---

### User Story 6 - Network Fault Tolerance with Exponential Backoff (Priority: P2)

As an operations engineer, I want the agent to reconnect to the backend using a predictable exponential backoff strategy after network failures, so that reconnection storms do not overwhelm the backend during large-scale outages.

**Why this priority**: Phase 1 mentioned "retry with backoff" but never specified the parameters, risking thundering herd problems at fleet scale.

**Independent Test**: Can be tested by blocking network access from the agent, observing the retry timing pattern (1s, 2s, 4s, 8s, ..., 60s cap), and confirming reconnection upon network restoration.

**Acceptance Scenarios**:

1. **Given** the agent loses its connection to the backend, **When** it begins reconnection attempts, **Then** it uses exponential backoff starting at 1 second, multiplying by 2.0 on each failure, capping at 60 seconds.
2. **Given** the agent has backed off to the 60-second maximum interval, **When** the backend becomes reachable again, **Then** the agent reconnects on its next attempt and immediately resumes telemetry streaming.
3. **Given** multiple agents lose connectivity simultaneously, **When** they begin reconnecting with backoff and jitter, **Then** no reconnection storm overwhelms the backend.

---

### User Story 7 - Protocol Payload Enforcement (Priority: P3)

As a backend engineer, I want a strict 4MB maximum payload size enforced on all telemetry batches, and a fully defined metadata schema for CSR requests, so that oversized messages cannot crash the ingestion pipeline and all identity metadata is consistently structured.

**Why this priority**: Without payload limits the backend is vulnerable to memory exhaustion from malformed or malicious payloads. The undefined CSR metadata schema caused inconsistencies in Phase 1 testing.

**Independent Test**: Can be tested by sending a batch exceeding 4MB and verifying rejection, and by submitting a CSR with all required metadata fields and verifying acceptance.

**Acceptance Scenarios**:

1. **Given** an agent prepares a telemetry batch, **When** the serialized payload exceeds 4MB, **Then** the agent splits it into multiple sub-batches each under 4MB before transmission.
2. **Given** a CSR request from the agent, **When** it is submitted, **Then** it includes the full metadata schema: VM hostname, VM unique identifier, organizational unit, and Subject Alternative Names (IP addresses and DNS names of the VM).
3. **Given** a CSR missing any required metadata field, **When** the backend receives it, **Then** it rejects the CSR with a descriptive validation error.

---

### Edge Cases

- What happens when the agent's filesystem becomes read-only? (The agent detects the condition, logs a critical error, and continues streaming telemetry directly without WAL buffering, accepting the risk of data loss.)
- What happens if the OTT expires before the agent can submit its CSR? (The provisioning system must detect the timeout and generate a new OTT for a retry.)
- What happens if the CRL grows excessively large? (CRL entries for certificates past their 30-day expiry are automatically pruned since the certificate itself is no longer valid.)
- How does the agent handle clock skew affecting certificate validation? (The agent must use the backend's timestamp from the certificate response as its reference for renewal calculations.)
- What happens when fsync fails due to disk I/O errors? (The agent logs the error, increments a health counter, and retries on the next batch cycle. After 3 consecutive fsync failures, it raises a critical health alert.)

## Requirements *(mandatory)*

### Functional Requirements

#### Security Hardening

- **FR-001**: The system MUST issue agent certificates with a 30-day Time-To-Live (TTL). The agent MUST initiate re-enrollment when its certificate is within 24 hours of expiration. Renewal CSRs MUST be authenticated using the agent's current, still-valid mTLS certificate (no new OTT required for renewals).
- **FR-002**: The system MUST implement a One-Time-Token (OTT) mechanism for CSR authorization. Each OTT MUST be generated during provisioning, bound to a single CSR, and invalidated immediately after use or upon a 10-minute expiration window.
- **FR-003**: The system MUST maintain a Certificate Revocation List (CRL). The backend MUST validate every incoming agent connection against the current CRL before accepting telemetry data.
- **FR-004**: All cryptographic key generation MUST use ECDSA P-256 (secp256r1). All private key files MUST be stored with 0600 file permissions (owner read/write only).

#### Infrastructure Resilience

- **FR-005**: The agent MUST self-enforce a resource envelope of 100MB maximum RAM usage and 10% maximum CPU utilization. When approaching limits, the agent MUST apply backpressure by reducing collection frequency or batch sizes.
- **FR-006**: The WAL MUST perform an fsync to disk after every 10th batch write to balance durability with performance.
- **FR-007**: When the WAL reaches its 50MB capacity limit, the agent MUST evict the oldest entries first ("drop-oldest" policy) to make room for new telemetry data.
- **FR-008**: On startup, the agent MUST validate WAL integrity. Corrupted trailing entries MUST be truncated and valid preceding data recovered. A fully unreadable WAL MUST be archived and replaced with a fresh file.

#### Protocol Rigor

- **FR-009**: The maximum serialized payload size per telemetry batch MUST NOT exceed 4MB. Batches exceeding this limit MUST be automatically split into compliant sub-batches before transmission.
- **FR-010**: The CSR metadata schema MUST include: VM hostname, VM unique identifier (machine-id), organizational unit, and Subject Alternative Names (containing the VM's IP addresses and DNS names).
- **FR-011**: The backend MUST reject any CSR that is missing required metadata fields and return a descriptive validation error.

#### Network Fault Tolerance

- **FR-012**: The agent MUST implement exponential backoff for reconnection with the following parameters: initial interval of 1 second, multiplier of 2.0, and maximum interval of 60 seconds. Random jitter MUST be applied to prevent thundering herd scenarios.

#### Observability

- **FR-013**: The agent and backend MUST emit structured log events for all security and recovery actions, including: OTT validation (success/failure), CRL revocation checks, certificate issuance and renewal, WAL recovery operations, and resource governor backpressure activations.
- **FR-014**: The agent MUST expose a lightweight health endpoint that reports key operational counters: OTT rejection count, CRL rejection count, WAL recovery count, WAL eviction count, backpressure activation count, current memory usage, and current CPU usage.

### Key Entities

- **Agent Certificate**: A time-limited identity credential with a 30-day TTL, issued via the CSR flow, containing the agent's public key, organizational metadata, and SANs.
- **One-Time-Token (OTT)**: A single-use authorization token with a 10-minute validity window, generated during provisioning and consumed during the initial CSR flow.
- **Certificate Revocation List (CRL)**: A backend-maintained list of revoked certificate serial numbers, checked on every agent connection attempt.
- **CSR Metadata**: The structured identity payload accompanying every certificate signing request, including hostname, machine-id, organizational unit, and SANs.
- **Resource Governor**: The agent-internal mechanism that monitors RAM and CPU consumption and applies backpressure when limits are approached.
- **WAL Recovery State**: The result of startup integrity validation—one of: clean (no issues), truncated (trailing corruption removed), or replaced (fully corrupt WAL archived and recreated).
- **Health Endpoint**: A lightweight local endpoint exposing operational counters for monitoring tools to scrape, covering security events, WAL state, and resource governor metrics.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of issued agent certificates have a TTL of exactly 30 days, and agents successfully re-enroll before expiration with zero telemetry gaps.
- **SC-002**: 100% of initial CSR flows are authorized by a valid OTT, and 100% of reused or expired OTTs are rejected.
- **SC-003**: Agent memory consumption remains below 100MB and CPU below 10% under sustained load across 95% of monitored intervals.
- **SC-004**: Following a simulated crash during WAL write, 100% of data committed before the last fsync point is recovered on restart.
- **SC-005**: Zero telemetry batches exceeding 4MB are accepted by the backend; all oversized batches are split before transmission.
- **SC-006**: Agent reconnection after network failure follows the defined backoff curve (1s → 2s → 4s → ... → 60s cap) with no reconnection storms observed at fleet scale (100+ agents).
- **SC-007**: A revoked agent certificate is rejected within 1 connection attempt after CRL update.
- **SC-008**: All hardening events (OTT validation, CRL checks, WAL recovery, backpressure activation) are observable via structured logs and queryable via the agent health endpoint within 5 seconds of occurrence.

## Assumptions

- The existing Phase 1 telemetry infrastructure (agent binary, CSR endpoint, WAL, gRPC streaming) is deployed and functional as the baseline for this hardening work.
- The backend Certificate Authority from Phase 1 supports configurable certificate TTLs and can be extended to maintain a CRL.
- The target VMs have hardware clocks accurate enough that a 24-hour renewal window is sufficient to prevent accidental expiry (NTP or equivalent is in use).
- The 50MB WAL limit from Phase 1 remains unchanged; this specification adds fsync semantics, eviction policy, and recovery to the existing WAL requirement.
- The fleet size for backoff/jitter validation is representative of enterprise deployments (100+ concurrent agents).
