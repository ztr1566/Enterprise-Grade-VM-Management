# Implementation Tasks: VM Telemetry Agent

**Feature**: VM Telemetry Agent
**Branch**: 002-v2-enterprise-upgrade
**Generated**: 2026-04-19

## Dependencies & Execution Order

1. **Phase 1 (Setup)**: Shared protocol contracts and configuration.
2. **Phase 2 (Foundational)**: Backend CA and core gRPC server.
3. **Phase 3 (User Story 2: Zero-Trust Provisioning)**: Agent identity lifecycle & deployment script.
4. **Phase 4 (User Story 1: Secure Batched Metrics)**: Hardware metric collection, WAL implementation, streaming.
5. **Phase 5 (User Story 3: Structured Batched Logs)**: Log aggregation and streaming.
6. **Phase 6 (Deprecation & Polish)**: Removing old SSH polling logic.

*Note: User Story 2 (Provisioning & Identity) is sequenced before User Story 1 (Metrics) because the identity and mTLS tunnel are prerequisites for transmitting any telemetry.*

## Phase 1: Setup & Protocol Contracts

**Goal**: Establish the shared gRPC definitions for identity and telemetry ingestion.

- [ ] T001 Initialize the `telemetry` protobuf module in `backend/internal/api/grpc/telemetry`.
- [ ] T002 [P] Compile `telemetry.proto` to generate Go structs and gRPC interfaces (`telemetry.pb.go` and `telemetry_grpc.pb.go`).
- [ ] T003 [P] Create initial directory structure for the agent daemon in `backend/cmd/agent` and `backend/internal/agent`.

## Phase 2: Foundational (Backend Core)

**Goal**: Establish the backend Certificate Authority (CA) and the unauthenticated gRPC listener capable of handling incoming CSRs.

- [ ] T004 Implement the internal CA logic in `backend/internal/ca/ca.go` capable of signing standard x509 CSRs.
- [ ] T005 Implement the `AgentIdentity` gRPC service handler in `backend/internal/api/grpc/telemetry/identity_handler.go`.
- [ ] T006 Configure the primary gRPC server in `backend/cmd/server/main.go` to expose the telemetry services over an mTLS-secured port, reserving the `SignCSR` method for unauthenticated bootstrap.

## Phase 3: Zero-Trust Automated Provisioning (User Story 2)

**Goal**: Deploy the Go agent binary via SSH, generate a local private key, and obtain an mTLS certificate via CSR without exposing private keys.
**Independent Test**: The script successfully deploys the binary, and the agent obtains its mTLS certificate from the backend without manual intervention.

- [ ] T007 [US2] Implement private key generation (ECDSA P-256) and CSR generation logic in `backend/internal/agent/bootstrap/csr.go`.
- [ ] T008 [US2] Create the agent's main bootstrap sequence in `backend/cmd/agent/main.go` that executes the CSR flow if no valid certificate exists locally.
- [ ] T009 [US2] Implement secure certificate storage (with 0600 file permissions for the key) in `backend/internal/agent/bootstrap/storage.go`.
- [ ] T010 [US2] Refactor `backend/scripts/provisioning/deploy_agent.sh` to remove legacy key deployment and exclusively SCP the agent binary and systemd service file.

## Phase 4: Secure Batched Metric Transmission & Resilience (User Story 1)

**Goal**: Collect hardware metrics, buffer them in a disk-backed WAL (up to 50MB), and stream them to the backend in batches.
**Independent Test**: The agent collects CPU/RAM data, buffers it, and correctly streams batched payloads to the backend's database.

- [ ] T011 [P] [US1] Implement hardware sampling logic (CPU, RAM, Disk, Network) in `backend/internal/agent/telemetry/metrics.go`.
- [ ] T012 [P] [US1] Implement the 50MB disk-backed Write-Ahead Log (WAL) structure and eviction policy in `backend/internal/agent/wal/disk_wal.go`.
- [ ] T013 [US1] Build the batched client-side streaming loop (500ms intervals) in `backend/internal/agent/telemetry/streamer.go`.
- [ ] T014 [US1] Implement the backend `StreamMetricsBatch` gRPC handler in `backend/internal/api/grpc/telemetry/metrics_handler.go`.
- [ ] T015 [US1] Wire the backend ingestion handler to the existing SQLite database in `backend/internal/db/sqlite.go`.

## Phase 5: Structured Batched Log Streaming (User Story 3)

**Goal**: Stream application and system logs in structured batched JSON formats to the backend.
**Independent Test**: Test log entries triggered on the VM are ingested into the backend via batched gRPC streams.

- [ ] T016 [US3] Implement log aggregation and JSON parsing logic in `backend/internal/agent/telemetry/logs.go`.
- [ ] T017 [US3] Integrate log payloads into the WAL implementation to share the 50MB buffer capacity.
- [ ] T018 [US3] Implement the backend `StreamLogsBatch` gRPC handler in `backend/internal/api/grpc/telemetry/logs_handler.go`.

## Phase 6: Deprecation & Polish

**Goal**: Remove legacy SSH polling mechanisms and verify system metrics.

- [ ] T019 Strip legacy SSH polling loops from `backend/internal/monitor/services.go` and `backend/internal/ssh/client.go`.
- [ ] T020 Create automated integration test verifying zero-trust CSR flow and batched metric ingestion in `backend/tests/integration/provisioning_test.go`.
- [ ] T021 Document exact 80% network overhead measurement methodology in `README.md`.