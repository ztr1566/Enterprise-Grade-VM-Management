---
description: "Task list for Telemetry Agent Hardening implementation"
---

# Tasks: Telemetry Agent Hardening

**Input**: Design documents from `/specs/004-telemetry-hardening/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Tests are included based on user request (Phase 5 Validation).
**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [x] T001 Create agent governance directory `backend/internal/agent/governance/`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T002 [P] Create `agent_tokens` table migration in `backend/internal/db/migrations/007_agent_tokens.sql`
- [x] T003 [P] Create `agent_certificates` table migration in `backend/internal/db/migrations/008_agent_certificates.sql`

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - mTLS Certificate Lifecycle Management (Priority: P1) 🎯 MVP

**Goal**: Implement 30-day mTLS certificate lifecycle, OTT-based CSR, and CRL revocation.

**Independent Test**: Provision an agent with an OTT, verify 30-day cert issuance, and verify rejection when added to CRL.

### Implementation for User Story 1

- [ ] T004 [P] [US1] Define OTT and Certificate data models in `backend/internal/models/models.go`
- [ ] T005 [P] [US1] Implement ECDSA P-256 key generation with 0600 permissions in `backend/internal/crypto/keys.go`
- [ ] T006 [US1] Update backend CSR endpoint to validate OTT in `backend/internal/api/grpc/telemetry/csr.go`
- [ ] T007 [P] [US1] Implement CRL validation gRPC interceptor in `backend/internal/api/grpc/interceptors/crl.go`
- [ ] T008 [US1] Implement agent cert auto-renewal logic via mTLS in `backend/internal/agent/telemetry/client.go`

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently

---

## Phase 4: User Story 2 - Resilient Disk-Backed Storage (Priority: P1)

**Goal**: Hardened disk-backed WAL with CRC32 checksums, fsync batches, and corruption recovery.

**Independent Test**: Write batches, manually corrupt the trailing bytes of the WAL file, and verify agent startup cleanly truncates the corrupted bytes.

### Implementation for User Story 2

- [ ] T009 [P] [US2] Update WAL entry format to include CRC32 checksums in `backend/internal/agent/wal/entry.go`
- [ ] T010 [US2] Implement fsync logic (every 10 batches) and 50MB eviction policy in `backend/internal/agent/wal/writer.go`
- [ ] T011 [US2] Implement truncating recovery scan on startup in `backend/internal/agent/wal/reader.go`

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently

---

## Phase 5: User Story 3 - Agent Resource Governance (Priority: P2)

**Goal**: Monitor agent RAM (100MB limit) and CPU (10% limit) to apply backpressure.

**Independent Test**: Query local HTTP endpoint and verify backpressure activates under simulated load.

### Implementation for User Story 3

- [ ] T012 [P] [US3] Implement internal ticker-based resource monitor in `backend/internal/agent/governance/monitor.go`
- [ ] T013 [P] [US3] Expose local HTTP health endpoint (`/health`) in `backend/internal/agent/telemetry/health.go`
- [ ] T014 [US3] Integrate monitor backpressure into telemetry collector in `backend/internal/agent/telemetry/collector.go`

**Checkpoint**: All user stories up to US3 should now be independently functional

---

## Phase 6: User Story 4 - High-Fidelity Streaming & Ingestion (Priority: P2)

**Goal**: Enforce 4MB max gRPC payload limits and exponential backoff.

**Independent Test**: Simulate network drops and observe agent 1s-60s backoff. Send large data and verify it is split before transmission.

### Implementation for User Story 4

- [ ] T015 [P] [US4] Implement 4MB payload enforcement interceptor in `backend/internal/api/grpc/interceptors/payload.go`
- [ ] T016 [US4] Update agent to measure and split large payloads in `backend/internal/agent/telemetry/sender.go`
- [ ] T017 [US4] Implement exponential backoff (1s-60s) with jitter in `backend/internal/agent/telemetry/connection.go`

**Checkpoint**: All user stories should now be independently functional

---

## Phase 7: Validation & Polish

**Purpose**: Execute specific failure scenarios to confirm checklists pass.

- [ ] T018 [P] Implement integration test for expired OTT and revoked cert failures in `backend/tests/integration/csr_test.go`
- [ ] T019 [P] Implement integration test for simulated WAL corruption and power loss in `backend/tests/integration/wal_test.go`
- [ ] T020 Run quickstart.md validation to ensure end-to-end agent operability

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3+)**: All depend on Foundational phase completion
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P2 → P3)
- **Validation (Final Phase)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P1)**: Can start after Foundational (Phase 2) - Independent
- **User Story 3 (P2)**: Can start after Foundational (Phase 2) - Independent
- **User Story 4 (P2)**: Can start after Foundational (Phase 2) - Independent

### Parallel Opportunities

- Foundational DB migrations (`T002`, `T003`) can run in parallel.
- Data models (`T004`), keys (`T005`), and CRL interceptor (`T007`) in US1 can run in parallel.
- Resource monitor (`T012`) and Health endpoint (`T013`) in US3 can run in parallel.
- Interceptor (`T015`) in US4 can be run in parallel with agent logic.
- Integration tests (`T018`, `T019`) can run in parallel.

---

## Parallel Example: User Story 3

```bash
# Launch independent component tasks together:
Task: "Implement internal ticker-based resource monitor in backend/internal/agent/governance/monitor.go"
Task: "Expose local HTTP health endpoint (/health) in backend/internal/agent/telemetry/health.go"
```

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: Test User Story 1 independently
5. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Deploy/Demo (MVP!)
3. Add User Story 2 → Test independently → Deploy/Demo
4. Add User Story 3 → Test independently → Deploy/Demo
5. Add User Story 4 → Test independently → Deploy/Demo
6. Each story adds value without breaking previous stories
