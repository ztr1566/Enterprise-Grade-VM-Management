---
description: "Detailed step-by-step task tracking for Web-Based VM Management Platform (Golang Architecture)"
---

# Tasks: Web-Based VM Management Platform

**Input**: Design documents from `/specs/001-vm-management/`
**Prerequisites**: plan.md, spec.md, data-model.md, contracts/, research.md

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story. Go backend tasks are sequentially prioritized before their React frontend counterparts.

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization, scaffolding for separated frontend (React) and backend (Go).

- [x] T001 Initialize the monorepo structure with `backend/` and `frontend/` directories
- [x] T002 [P] Initialize Golang backend layout (`go mod init`, `cmd/server/main.go`) inside `backend/`
- [x] T003 [P] Initialize React Vite frontend with TypeScript, Tailwind CSS, and React Router inside `frontend/`
- [x] T004 [P] Configure global linting (golangci-lint / ESLint) and Prettier formats for both workspaces

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure, database connections, security gates, and audit logging that MUST be complete before user stories. Backend implemented first.

- [x] T005 Setup SQLite database connection utilizing `mattn/go-sqlite3` in `backend/internal/db/sqlite.go`
- [x] T006 [P] Implement AES-256-GCM Crypto service using Go's `crypto/cipher` for secure DB parameter storage in `backend/internal/crypto/aes.go`
- [x] T007 Implement structured JSON Audit Logging module utilizing `uber-go/zap` in `backend/internal/audit/logger.go`
- [x] T008 [P] Implement raw WebSocket Upgrader logic utilizing `gorilla/websocket` in `backend/internal/ws/router.go`
- [x] T009 Implement centralized Go error handling and HTTP response formatter for API routes in `backend/internal/api/errors.go`
- [x] T010 Implement RBAC/Authentication HTTP middleware (JWT validation) in `backend/internal/api/middleware/auth.go`

**Checkpoint**: Foundation, auditing, and security perimeters ready - user story implementations can begin.

---

## Phase 3: User Story 1 - Centralized VM Inventory Management (Priority: P1) 🎯 MVP

**Goal**: System administrators can securely authenticate, add, tag, and list Virtual Machines.
**Independent Test**: Can log into the frontend, add a server with tags, observe AES-encrypted keys written accurately into SQLite, and verify Audit Log Entries are generated for credential access.

### Implementation



**Checkpoint**: At this point, User Story 1 is fully functional. The dashboard is protected via RBAC, securely adds VMs, generates audit trails, and manages tags.

---

## Phase 4: User Story 2 - Interactive Web Terminal (Priority: P1)

**Goal**: Full Bash terminal directly from the browser utilizing Go's native standard library concurrency over WebSockets.
**Independent Test**: Trigger a connection to an active VM and verify input mapping across Gorilla WebSockets directly to the `crypto/ssh` session pipeline.

### Implementation

- [x] T018 [US2] Implement core `golang.org/x/crypto/ssh` connection manager with strict timeout limits in `backend/internal/ssh/client.go`
- [x] T019 [US2] Link the WebSocket `terminal.*` event struct parser to `ssh.Session.StdinPipe/StdoutPipe` in `backend/internal/ws/terminal.go`
- [x] T020 [P] [US2] Setup base `xterm.js` Component with Tailwind integrations in `frontend/src/components/TerminalView.tsx`
- [x] T021 [US2] Implement bridge using `xterm-addon-attach` strictly transmitting raw strings via standard WebSockets in `frontend/src/services/wsClient.ts`
- [x] T022 [US2] Integrate the `TerminalView` into the `Inventory` page (triggering an SSH handshake when a VM is clicked) in `frontend/src/pages/Inventory.tsx`

**Checkpoint**: A fully interactive, responsive terminal overlay is actively piped through the Go streaming server.

---

## Phase 5: User Story 3 - Live Telemetry Dashboard (Priority: P2)

**Goal**: View real-time system metrics via agentless SSH collection.
**Independent Test**: Assure `goroutines` appropriately dispatch `top`/`/proc/stat` fetches and JSON marshals the results without memory leaks.

### Implementation

- [x] T023 [US3] Implement agentless telemetry extraction by launching detached `goroutines` running periodic short-lived `crypto/ssh` calls in `backend/internal/telemetry/extractor.go`
- [x] T024 [US3] Broadcast parsed `telemetry.update` Go structures via `gorilla/websocket` hub iterator in `backend/internal/ws/telemetry.go`
- [x] T025 [P] [US3] Build dynamic React Chart overlays parsing incoming WS JSON in `frontend/src/components/TelemetryGraphs.tsx`
- [x] T026 [US3] Tie frontend charts into the state tree mapped directly to the active WebSocket session inside `frontend/src/pages/Dashboard.tsx`

**Checkpoint**: Core operational metrics are continuously streaming via background Go routines.

---

## Phase 6: User Story 4 - Application & Service Monitoring (Priority: P2)

**Goal**: Track the systemd hook status of specifically deployed target daemons autonomously.
**Independent Test**: Trigger an `nginx` crash remotely and verify the WS payload toggles the React component status circle from green to red.

### Implementation

- [x] T027 [US4] Create schema mapping for `services` attached by `vm_id` in `backend/internal/db/migrations/002_services.sql`
- [x] T028 [US4] Implement Go models and HTTP getter route `GET /api/vms/:id/services` in `backend/internal/api/handlers/service.go`
- [x] T029 [US4] Expand the background telemetry `goroutines` to inject `systemctl is-active [service]` checks inside `backend/internal/telemetry/extractor.go`
- [x] T030 [US4] Distribute state diffs via `service.health` WebSocket broadcast in `backend/internal/ws/telemetry.go`
- [x] T031 [P] [US4] Create Service tracker UI lists visually changing states based on active WebSocket hooks in `frontend/src/components/ServiceTracker.tsx`

**Checkpoint**: Crucial daemon monitoring occurs seamlessly entirely server-side.

---

## Phase 7: User Story 5 - Quick-Action Control Panel (Priority: P3)

**Goal**: Provide HTTP action overrides to manually force daemon status executions without terminal access.
**Independent Test**: Select "Restart" internally calling `systemctl` successfully passing state responses over HTTP boundaries.

### Implementation

- [x] T032 [US5] Extend the `ssh/client.go` package to support synchronous one-shot bash command executions directly returning byte outputs inside `backend/internal/ssh/executor.go`
- [x] T033 [US5] Implement HTTP handler `POST /api/vms/:id/actions` strictly bound to the `ActionType` enum in `backend/internal/api/handlers/action.go`
- [x] T034 [P] [US5] Create Action buttons featuring optimistic UI loaders and inline error parsing in `frontend/src/components/ActionControls.tsx`

**Checkpoint**: Granular lifecycle operations complete dynamically.

---

## Phase 8: User Story 6 - Real-time Log Observation (Priority: P2)

**Goal**: Administrators can view live-streamed systemd logs directly via WebSockets to diagnose service failures without SSH access.
**Independent Test**: Trigger a service restart and observe the new log entries appearing in the UI console without refreshing.

### Implementation

- [x] T035 [US6] Implement WebSocket log streaming handler in `backend/internal/monitor/logs.go` using `journalctl -f`.
- [x] T036 [US6] Add `/api/vms/:id/logs/:service` WebSocket route to the main router.
- [x] T037 [P] [US6] Create `LogViewer.tsx` component with terminal-style output and auto-scroll logic in `frontend/src/components/LogViewer.tsx`.
- [x] T038 [US6] Integrate `LogViewer` as a sub-tab in the `VMDetails` view.

---

## Phase 9: Validation & Testing (Success Criteria Verification)

**Goal**: Load testing and performance validation ensuring full compliance with the 5 measurable Success Criteria (SC).
**Independent Test**: Test suite runs and passes asserting all thresholds cleanly.

### Implementation

- [x] T039 Write Go benchmark tests validating that adding a new target VM and querying its data takes <60s per SC-001 in `backend/internal/api/handlers/vm_test.go`
- [x] T040 Write Go integration tests validating telemetry metric collection refreshes exactly every 5 seconds per SC-002 in `backend/internal/telemetry/extractor_test.go`
- [x] T041 Develop `k6` load testing script to validate the Web Terminal keystroke latency is under 150ms on stable connections per SC-003 in `backend/tests/load/terminal_latency.js`
- [x] T042 Write automated frontend/backend integration tests to ensure Quick Action commands execute within 3 seconds per SC-004 in `backend/tests/integration/actions_test.go`
- [x] T043 Develop `k6` load testing script to simulate 50 concurrent target VM monitoring and terminal WebSocket sessions simultaneously per SC-005 in `backend/tests/load/concurrency_scale.js`

---

## Final Phase: Polish & Cross-Cutting Concerns

**Purpose**: Fix boundary exceptions specifically noted in spec analytics and handle environment transitions gracefully.

- [x] T044 Implement explicit exponential backoff WebSocket reconnection handler with persistent visual state overlay in `frontend/src/utils/wsBackoff.ts` and `frontend/src/components/ConnectionOverlay.tsx`
- [x] T045 Standardize styling across the Dashboard UI verifying strict compliance with Tailwind 'dark:' variants.
- [x] T046 Finalize `README.md` and Go `Makefile` for streamlined developer execution paths in `backend/Makefile`

---

## Dependencies & Execution Order

### Phase Dependencies
- **Phase 1 (Setup)**: Immediate parallel
- **Phase 2 (Foundational)**: Blocks US1. Sets up the SQLite hooks, AES encryption buffers, Zap audit logging, and JWT middleware required to safely store hosts.
- **Phase 3 (US1)**: Directly prerequisites the remaining User Stories via RBAC token issuances and physical record IDs in SQLite.
- **Phase 4 (US2)** -> **Phase 7 (US5)**: Implemented in consecutive priority tiers. 
- **Phase 8 (Testing)**: Validates SC criteria explicitly. Must be executed after Phase 7.

### Parallel Opportunities
- Initialization of Go environments (T002) and Vite wrappers (T003).
- Setup of database logic (T005), Crypto AES keys (T006), and raw Gorilla Websocket initializers (T008).
- During US1, API CRUD bindings and Audit bindings in Go (T012, T013, T014) securely operate concurrently decoupled from UX components mapping JWT boundaries in React (T015, T016).
- Building visual React chart handlers (T025) isolates identically from internal Go extraction routines (T023).
