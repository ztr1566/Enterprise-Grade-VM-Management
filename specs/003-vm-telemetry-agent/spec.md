# Feature Specification: VM Telemetry Agent

**Feature Branch**: `[###-vm-telemetry-agent]`  
**Created**: April 19, 2026  
**Status**: Draft  
**Input**: User description: "Update the existing spec.md for the VM Telemetry Agent to correct critical architectural deficiencies in the current draft. Integrate the following technical mandates into the Requirements and User Stories: Telemetry Transmission (Performance): Replace continuous streaming of individual log lines with batched client-side streaming for both metrics and structured logs (e.g., 500ms intervals or 100-event batches) to mitigate network overhead. Local Buffering Resilience (Data Integrity): Upgrade the local buffer requirement (FR-009) to explicitly mandate a disk-backed Write-Ahead Log (WAL). This ensures data integrity during sudden power loss or kernel panics, capped at a 50MB limit. Security & Identity Lifecycle (Zero-Trust): Modify the automated provisioning requirement (FR-001). The initial SSH bootstrap must only copy the agent binary. The agent must then independently generate its own private key and execute a Certificate Signing Request (CSR) flow to the backend to obtain its mTLS certificate. Private keys must never traverse the network via SSH."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Secure Batched Metric Transmission (Priority: P1)

As a system administrator, I want the VM metrics (CPU, RAM, Disk, Network) to be pushed from a dedicated agent to the backend in efficient batches securely via gRPC and mTLS, so that I can reliably monitor my VM fleet with high performance and minimal network overhead.

**Why this priority**: Continuous observability is critical for operations, and migrating away from SSH polling improves both security and scalability. Batching prevents network saturation.

**Independent Test**: Can be independently tested by running a mock agent that pushes batched CPU/RAM data to the backend endpoint and observing that the metrics are correctly ingested and stored in the database.

**Acceptance Scenarios**:

1. **Given** a provisioned Go agent running on a VM, **When** it collects periodic system metrics, **Then** it successfully streams them to the backend gRPC server in batched payloads (e.g., 500ms intervals or 100 events) over mTLS.
2. **Given** an invalid or expired mTLS certificate, **When** the agent attempts to push metrics, **Then** the backend rejects the connection and logs an authorization failure.

---

### User Story 2 - Zero-Trust Automated Provisioning (Priority: P1)

As a DevOps engineer, I want the provisioning process to automatically deploy the Go agent binary via SSH, after which the agent independently generates its private key and obtains a certificate via a CSR flow, so that new VMs are securely integrated without private keys ever traversing the network.

**Why this priority**: Manual installation is unscalable. Zero-trust identity lifecycle ensures robust security and avoids distributing private keys over SSH.

**Independent Test**: Can be verified by running the provisioning script on a fresh Linux VM and validating that only the agent binary is transferred via SSH, and subsequently observing the agent generating its key and receiving a valid certificate from the backend.

**Acceptance Scenarios**:

1. **Given** a new Standard Linux or Containerized VM, **When** the backend initiates the setup process, **Then** it automatically SCPs only the agent binary to the target VM and starts the service.
2. **Given** the agent is running for the first time, **When** it initializes, **Then** it generates a private key locally and successfully completes a CSR flow with the backend to receive its mTLS certificate.

---

### User Story 3 - Structured Batched Log Streaming (Priority: P2)

As a developer troubleshooting a VM, I want the agent to stream application and system logs in batched, structured JSON format (with timestamp, severity, component, and message) to the backend, so that I can easily query and filter log data without overwhelming the network.

**Why this priority**: Structured logs improve debuggability over raw text, but base metrics are a higher operational priority. Batching optimizes network performance.

**Independent Test**: Can be tested by triggering test log entries on the VM and verifying that the backend ingests them in batches with the correct parsed fields.

**Acceptance Scenarios**:

1. **Given** a high volume of log events on the VM, **When** the agent captures them, **Then** it transmits the logs to the backend via gRPC in batched payloads formatted as structured JSON.

---

### Edge Cases

- What happens when the backend gRPC server is temporarily unreachable? (The agent should buffer logs/metrics locally in a disk-backed Write-Ahead Log (WAL) up to a 50MB limit and retry with backoff).
- How does the system handle an agent on an unsupported OS? (The provisioning script should fail fast and report an incompatibility error).
- What happens if the VM experiences sudden power loss or kernel panic? (Data integrity is maintained up to the last synced event in the disk-backed WAL).
- What happens if the disk-backed WAL reaches its 50MB limit? (The agent should prioritize keeping itself alive and discard the oldest telemetry).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST provide an automated provisioning script that uses SSH to deploy only the Go agent binary to target VMs. The agent MUST independently generate its private key and execute a Certificate Signing Request (CSR) flow to the backend to obtain its mTLS certificate. Private keys MUST NEVER traverse the network via SSH.
- **FR-002**: The agent MUST support deployment on Standard Linux (Ubuntu/Debian, RHEL/CentOS) and Containerized Workloads.
- **FR-003**: The agent MUST collect CPU, RAM, Disk, and Network metrics natively.
- **FR-004**: The agent MUST push collected metrics to the backend in batched client-side streams (e.g., 500ms intervals or 100-event batches) via a secure gRPC connection using mTLS.
- **FR-005**: The agent MUST capture and stream system/application logs to the backend in batched payloads formatted as structured JSON containing timestamp, severity, component, and message.
- **FR-006**: The backend MUST expose a gRPC endpoint configured to process CSRs and validate client mTLS certificates before accepting telemetry data.
- **FR-007**: The system MUST deprecate and safely disable the existing SSH-based agentless polling for metrics and logs.
- **FR-008**: The backend MUST retain SSH access capabilities exclusively for direct terminal access and explicit administrative commands.
- **FR-009**: The agent MUST implement a local disk-backed Write-Ahead Log (WAL) to queue metrics and logs when the connection to the backend is lost, ensuring data integrity during sudden power loss or kernel panics. The WAL MUST be capped at a 50MB limit, after which the oldest telemetry is discarded.

### Key Entities

- **VM Agent**: The Go daemon installed on the VM responsible for telemetry collection, batched transmission, and secure identity lifecycle (key generation & CSR).
- **Metric Payload**: A structured message containing a batch of CPU, RAM, Disk, and Network telemetry data.
- **Log Payload**: A structured JSON payload representing a batch of log lines with fields for timestamp, severity, component, and message.
- **Agent Identity**: The independently generated private key and resulting mTLS client certificate uniquely identifying a specific VM agent to the backend.
- **Write-Ahead Log (WAL)**: The disk-backed buffer ensuring telemetry data integrity during connectivity losses or sudden system crashes.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of telemetry (metrics and logs) is transmitted over the new batched streaming architecture instead of SSH polling.
- **SC-002**: CPU usage on the backend server related to polling drops by at least 50% compared to the legacy polling method.
- **SC-003**: The automated provisioning script successfully deploys the agent binary and completes the zero-trust CSR flow on 95% of standard Linux instances in under 60 seconds.
- **SC-004**: Network overhead overhead per event is reduced by at least 80% due to the introduction of batched telemetry streaming.

## Assumptions

- Target VMs have sufficient outbound network access to reach the backend endpoints for both CSR flows and telemetry streaming.
- The backend already has a certificate authority (CA) capable of securely signing agent CSRs.
- VMs have write access to a local filesystem to persist the 50MB disk-backed Write-Ahead Log (WAL) and the locally generated private key.