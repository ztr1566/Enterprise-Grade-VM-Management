---
description: "Detailed step-by-step task tracking for V2 Enterprise Upgrade"
---

# Tasks: V2 Enterprise Upgrade

**Input**: Design documents from `/specs/002-v2-enterprise-upgrade/`
**Prerequisites**: plan.md, spec.md, data-model.md, contracts/, research.md, quickstart.md

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story. Backend tasks are sequentially prioritized before their frontend counterparts.

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project scaffolding for V2 — new directories, dependencies, and database migration.

- [x] T001 Create the provisioning script directory structure at `backend/scripts/provisioning/`
- [x] T002 [P] Install `react-router-dom` dependency in `frontend/` for multi-page navigation
- [x] T003 [P] Create database migration file `backend/internal/db/migrations/003_provisioning.sql` with the `provisioning_records` table schema per data-model.md

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core provisioning infrastructure and database wiring that MUST be complete before any user story can be implemented.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [x] T004 Create the provisioning shell script at `backend/scripts/provisioning/setup_vm_sudoers.sh` implementing: accept username as $1, generate scoped NOPASSWD rules for `/usr/bin/systemctl *` and `/usr/bin/journalctl *` only, write to temp file, validate with `visudo -c -f`, move to `/etc/sudoers.d/vm-platform-<username>` with `chmod 0440` on success, remove temp file and exit 1 on failure, self-cleanup of the uploaded script
- [x] T005 Create the Provisioning Record model with CRUD operations (CreateRecord, GetByVMID, UpdateStatus) in `backend/internal/models/provisioning.go`
- [x] T006 Apply migration `003_provisioning.sql` in the server startup sequence in `backend/cmd/server/main.go`

**Checkpoint**: Foundation ready — provisioning script exists, DB schema is live, model CRUD is available.

---

## Phase 3: User Story 1 — Zero-Touch Secure Provisioning (Priority: P1) 🎯 MVP

**Goal**: Newly registered VMs are automatically provisioned with scoped sudoers rules. No sudo passwords are ever handled by the platform.

### Implementation

- [x] T007 [US1] Create the Go provisioner engine at `backend/internal/ssh/provisioner.go` using `//go:embed` to bundle `internal/ssh/provisioning/setup_vm_sudoers.sh`, implementing `ProvisionVM(db *sql.DB, vmID string) error` that: fetches the VM record, sets provisioning status to "pending", opens an SSH session via existing `NewSession()`, uploads the embedded script to `/tmp/vm-platform-provision.sh` via heredoc, executes `sudo bash /tmp/vm-platform-provision.sh <username>`, updates provisioning status to "provisioned" or "failed" based on exit code, and logs audit events (`VM_PROVISIONING_START`, `VM_PROVISIONED` or `VM_PROVISIONING_FAILED`)
- [x] T008 [US1] Modify `CreateVM` handler in `backend/internal/api/handlers/vm.go` to: create a `provisioning_records` entry with status "pending" after persisting the VM, launch `go ssh.ProvisionVM(h.DB, vm.ID)` asynchronously, and include `provisioning_status: "pending"` in the 201 response
- [x] T009 [US1] Create the provisioning API handler at `backend/internal/api/handlers/provisioning.go` implementing: `POST /api/vms/{id}/provision` (re-provision trigger) and `GET /api/vms/{id}/provisioning` (status query) per contracts/api.md
- [x] T010 [US1] Register the new provisioning routes (`POST /api/vms/{id}/provision` and `GET /api/vms/{id}/provisioning`) in `backend/cmd/server/main.go` behind `AuthMiddleware`
- [x] T011 [P] [US1] Create the `ProvisioningBadge.tsx` component in `frontend/src/components/ProvisioningBadge.tsx` displaying status badges: green "Provisioned", red "Failed" with retry button, yellow "Pending" with spinner, grey "Not Provisioned"
- [x] T012 [US1] Update `frontend/src/pages/Inventory.tsx` to: fetch provisioning status alongside VM list, render `ProvisioningBadge` for each VM row, and wire the retry button to `POST /api/vms/{id}/provision`

---

## Phase 3.5: Identity & Multi-User Refactor (Enterprise Hardening)

**Goal**: Separate high-privilege management credentials from scoped user access and support provisioning for multiple OS users.

- [x] T012a [Hardening] Refactor database schema (migration 005) to rename `username` to `management_username` for clarity of privilege.
- [x] T012b [Hardening] Update `backend/internal/models/models.go` and handlers to support the `management_username` field.
- [x] T012c [Hardening] Create `vm_access` table (migration 004) to manage multi-user OS access per VM.
- [x] T012d [Hardening] Update `ProvisionVM` engine to fetch all authorized users from `vm_access` and provision them in a loop, with fallback to `management_username`.
- [x] [US1] Fix wildcard matching in `setup_vm_sudoers.sh` to explicitly allow both base and wildcard command versions.
- [x] [US1] Enhance `NewSession` and `TerminalHandler` to support optional `login_as` parameter for user-specific SSH sessions.
- [x] [P] Update `AddVmModal.tsx` to label management username as "High Privilege" for clarity.
- [x] [Fix] Resolve compilation errors in `main.go` and `vm.go` caused by refactoring side effects.
- [x] [Fix] Update `003_provisioning.sql` to include `created_at` DATETIME column and migrate existing records.
- [x] [Fix] Refactor `vm_test.go` to use `management_username` and apply all migrations in `setupTestDB`.
- [x] [Fix] Add explicit `zap.Error` logging in `handlers/provisioning.go` before HTTP errors to prevent silent failures.
- [x] [Fix] Validate safe empty-state handling in `GetUsersByVM` and `ProvisionVM` fallback logic.

**Checkpoint**: Identity architecture is hardened. Multi-user provisioning is operational.

---

## Phase 4: User Story 2 — Multi-Page Dashboard Architecture (Priority: P1)

**Goal**: Replace the single-page modal UI with a multi-page layout using React Router with persistent sidebar navigation and deep linking.

### Implementation

- [x] T013 [US2] Create the `Sidebar.tsx` component in `frontend/src/components/Sidebar.tsx` rendering a persistent navigation sidebar with links to: Dashboard/Overview (`/`), Inventory (`/inventory`), Audit Log (`/audit`), and Settings (`/settings`), with active-state highlighting
- [x] T014 [US2] Refactor `frontend/src/App.tsx` to use React Router v6 with a layout wrapper that renders `Sidebar` persistently alongside route-specific content via `<Outlet />`
- [x] T015 [US2] Create the `VMDetailPage.tsx` at `frontend/src/pages/VMDetailPage.tsx` as a full-page VM detail view at route `/vms/:id` with tabbed navigation for: Metrics, Services, Logs, Network, and Processes
- [x] T016 [US2] Create the `AuditLogPage.tsx` at `frontend/src/pages/AuditLogPage.tsx` implementing paginated audit log display at route `/audit` using `GET /api/audit` per contracts/api.md
- [x] T017 [US2] Create the audit log API handler at `backend/internal/api/handlers/audit.go` implementing `GET /api/audit` with pagination (page, limit query params) and register the route in `backend/cmd/server/main.go`
- [x] T018 [P] [US2] Create the `SettingsPage.tsx` at `frontend/src/pages/SettingsPage.tsx` as a placeholder settings page at route `/settings`
- [x] T019 [US2] Update `frontend/src/pages/Inventory.tsx` to navigate to `/vms/{id}` on VM row click instead of opening a modal/panel

**Checkpoint**: Full multi-page navigation is operational. All pages are deep-linkable. Browser history works correctly.

### Phase 4.5: Enterprise UI Overhaul & Logic Hardening

- [x] [Fix] Verified `pinger.go` and SSH `client.go` correctly use `management_username` — no legacy `username` references remain.
- [x] [Fix] Created `backend/internal/api/handlers/dashboard.go` with `GET /api/dashboard/stats` endpoint querying aggregate VM, provisioning, and audit counts.
- [x] [Fix] Added `GetVMStats` handler to `vm.go`, importing `monitor` package. Uncommented `/api/vms/{id}/stats` route in `main.go`.
- [x] [Fix] Registered `DashboardHandler` and `/api/dashboard/stats` route in `cmd/server/main.go`.
- [x] [Fix] Added `AuditLog` struct and `GetAuditLogs` paginated query to `models.go`.
- [x] [UI] Redesigned `DashboardView.tsx` with live data from `/api/dashboard/stats`, hero cards with Lucide icons and percentage indicators.
- [x] [UI] Redesigned `Sidebar.tsx` — compact, dark-themed, improved active-state with accent glow.
- [x] [UI] Redesigned `Layout.tsx` — dynamic page title, sign-out button, CSS page transition via key prop.
- [x] [UI] Redesigned `VMDetailView.tsx` — tabbed navigation with live SSH metrics panel (CPU/RAM/Disk bars).
- [x] [UI] Redesigned `AuditLogPage.tsx` — dark table theme, color-coded event badges, pagination controls.
- [x] [UI] Redesigned `SettingsPage.tsx` — dark card grid layout with Lucide icons.
- [x] [UI] Rewrote `index.css` with full enterprise design system: CSS variables, page transitions, hero cards, status badges with glow, data tables, metric bars.

**Checkpoint**: Enterprise-grade dark/slate UI is operational. Dashboard shows live data. Metrics pipeline is functional.

### Phase 4.6: Data-Bridge and Stats Logic Hardening
- [x] [Backend] Updated SQL queries in `GetDashboardStats` to correctly map `not_provisioned` to the `Pending` UI state, resolving the 0 metrics issue.
- [x] [Backend] Verified struct tags in `models.go` and `vm.go` use `json:"management_username"` matching frontend expectations.
- [x] [Backend] Refactored `pinger.go` to use `sshclient.NewSession` instead of a TCP probe, meaning it now strictly uses `management_username` for true connectivity verification.
- [x] [Frontend] Ensured `DashboardView.tsx` auto-fetches using `useEffect` immediately on mount, and `VMDetailView.tsx` auto-fetches when navigating to the Metrics tab.
- [x] [Frontend] Added robust error handling in `VMDetailView.tsx` to display specific API/SSH error messages in a red alert box instead of a generic failure log.

### Phase 4.7: UI Clutter and AppBar Elevation
- [x] [UI] Redesigned `Layout.tsx` to elevate the search bar, server count badge, Key Vault quick link, and a new user 'A' avatar with a Logout dropdown menu to the global AppBar.
- [x] [UI] Removed the redundant, secondary internal toolbars from `Inventory.tsx` and `KeysManagement.tsx`.
- [x] [UI] Refined `Inventory.tsx` header layout and upgraded VM cards to use the `premium-card` styling for a sleeker grid appearance.
- [x] [UI] Cleaned up `KeysManagement.tsx` header and simplified the SSH Key card design, resolving all UI clutter across main views.
- [x] [Fix] Resolved JSX syntax error in `Inventory.tsx` by correctly closing conditional rendering blocks and consolidating imports.

### Phase 4.8: Multi-User Access & CSS Standardization
- [x] [Backend] Updated Terminal WebSocket handler and `sshclient.NewSession` to support `login_as` parameter for user-specific sessions.
- [x] [Backend] Implemented `GET /api/vms/{id}/users` and `POST /api/vms/{id}/users` endpoints.
- [x] [Backend] Updated `VM` model to include `AuthorizedUsers` for streamlined frontend integration.
- [x] [Backend] Prepared `VMStats` struct with `net_in` and `net_out` placeholders for Phase 5 network telemetry.
- [x] [Frontend] Implemented "Access" tab in `VMDetailView.tsx` with `AccessManager` component for OS user management.
- [x] [Frontend] Updated `Inventory.tsx` Terminal button with a dropdown menu for login user selection.
- [x] [UI] Standardized CSS by moving `@import` to the top of `index.css` and adding shared `.dropdown-menu` components.
- [x] [UI] Verified uniform VM card heights and consistent grid layout across inventory views.
- [x] [Backend/Frontend] Implemented User Access Revocation (De-provisioning) with DB removal, SSH sudoers cleanup, and confirmation UI.

---

## Phase 5: User Story 3 — Advanced Network Telemetry (Priority: P2)

**Goal**: Display real-time network throughput (bytes in/out per second) and active TCP connection count for each VM.

### Implementation

- [x] T020 [US3] Create the network metrics collector at `backend/internal/monitor/network.go` implementing: SSH command to read `/proc/net/dev` for byte counters (filtering out `lo`, `docker0`, `veth*`, `br-*`), SSH command `ss -tun state established | wc -l` for active TCP count, rate calculation by diffing consecutive samples, and a `NetworkStats` struct
- [x] T021 [US3] Extend the `GetVMStats` handler in `backend/internal/api/handlers/vm.go` to include `network` object (rx_bytes_sec, tx_bytes_sec, active_connections) in the stats response per contracts/api.md
- [x] T022 [P] [US3] Create `NetworkChart.tsx` at `frontend/src/components/NetworkChart.tsx` displaying a real-time dual-axis time-series chart (RX/TX bytes per second) using Recharts, with an active TCP connections counter badge
- [x] T023 [US3] Integrate `NetworkChart` into the Network tab of `VMDetailPage.tsx` in `frontend/src/pages/VMDetailPage.tsx`, polling `GET /api/vms/{id}/stats` every 5 seconds

**Checkpoint**: Network telemetry is live on each VM's detail page. Charts update every 5 seconds.

---

## Phase 6: User Story 4 — Process Explorer (Priority: P2)

**Goal**: Display a live, sortable table of the top 20 processes on a VM, ranked by CPU or memory usage.

### Implementation

- [x] T024 [US4] Extend the metrics collector in `backend/internal/monitor/metrics.go` to add a `FetchProcesses` function that executes `ps aux --sort=-%cpu | head -21` via SSH and parses output into a `[]ProcessInfo` struct (PID, user, cpu_pct, mem_pct, command)
- [x] T025 [US4] Include the `processes` array in the `GetVMStats` response in `backend/internal/api/handlers/vm.go` per contracts/api.md
- [x] T026 [P] [US4] Create `ProcessTable.tsx` at `frontend/src/components/ProcessTable.tsx` displaying a sortable table of processes with columns: PID, User, CPU%, MEM%, Command — with client-side sort toggling on column header click
- [x] T027 [US4] Integrate `ProcessTable` into the Processes tab of `VMDetailPage.tsx` in `frontend/src/pages/VMDetailPage.tsx`, auto-refreshing every 5 seconds

### Phase 6.5: Interactive Process Management
- [x] [Backend] Implemented `POST /api/vms/{id}/processes/{pid}/signal` with support for KILL, TERM, and NICE.
- [x] [Backend] Optimized `FetchProcesses` with `ps -eo` for cleaner parsing and reliable PID mapping.
- [x] [Frontend] Enhanced `ProcessTable` with Action buttons (Kill, Reload) and a priority (Nice) selector.
- [x] [Frontend] Updated telemetry polling to 2 seconds for high-precision process monitoring.

**Checkpoint**: Process explorer shows the top 20 processes, sortable by CPU or memory. Interactive actions (Kill, Reload, Nice) are fully functional.

### Phase 6.6: Process Search & Stability (Refined)
- [x] [Fix] Added missing `Logger` field to `VMHandler` and injected it via `NewVMHandler` in `main.go`.
- [x] [Fix] Added missing `strings` import in `vm.go`.
- [x] [Backend] Fixed `ps` command to use `pcpu` and `pmem` for broader Linux compatibility.
- [x] [Backend] Bound search keyword specifically to the `COMMAND` field using `awk` with `tolower` for case-insensitivity.
- [x] [Backend] Quieted high-frequency polling logs by only logging non-empty search queries.
- [x] [Frontend] Added search input field to `ProcessTable` with real-time "Searching..." indicators and empty state messaging.
- [x] [Frontend] Fixed fetch logic to use `encodeURIComponent` for search queries.

### Phase 6.7: Systemd Service Management
- [x] [Backend] Created `internal/monitor/services.go` to execute and parse `systemctl list-units`.
- [x] [Backend] Extended `sshclient.RunSystemdCommand` to support `enable` and `disable` actions.
- [x] [Frontend] Created `ServiceTable.tsx` with search, active/inactive badges, and action buttons.
- [x] [Frontend] Integrated `ServiceTable` into `VMDetailView.tsx` with a dedicated 5-second polling loop.
- [x] [Polish] Improved Services UI with fixed layout, system service toggle, and enhanced status badges.

### Phase 6.8: VM Power Management
- [x] [Backend] Updated sudoers provisioning script (`setup_vm_sudoers.sh`) and reprovisioned all existing VMs to allow passwordless execution of power commands (`reboot`, `poweroff`, `systemctl suspend`).
- [x] [Backend] Established the `POST /api/vms/{id}/power` endpoint and mapped actions to specific absolute binaries.
- [x] [Frontend] Added a unified Lucide `Power` dropdown component across both `Inventory.tsx` and `VMDetailView.tsx` headers.

### Phase 7: Firewall & SELinux Management
- [x] [Backend] Updated sudoers provisioning script to whitelist `firewall-cmd`, `ufw`, `setenforce`, `getenforce`, and `sestatus`.
- [x] [Backend] Re-provisioned all reachable VMs with updated sudoers rules.
- [x] [Backend] Created `internal/api/handlers/security.go` with `GET /security`, `POST /firewall/rules`, and `POST /selinux` endpoints.
- [x] [Backend] Enhanced security detection to return `NOT_INSTALLED` if binaries are missing.
- [x] [Backend] Added `POST /api/vms/{id}/security/install` for automated component installation (firewalld/SELinux).
- [x] [Backend] Auto-detects firewall backend (firewalld vs ufw) and parses rules accordingly.
- [x] [Backend] SELinux status detection via `sestatus`/`getenforce` with runtime mode toggling via `setenforce`.
- [x] [Backend] All firewall rule modifications and SELinux mode changes are audited.
- [x] [Frontend] Created `SecurityPanel.tsx` with Firewall Control (rules table + add/remove form) and SELinux Dashboard (mode toggle).
- [x] [Frontend] Added "Install & Configure" buttons for missing security components with enterprise styling.
- [x] [Frontend] Added SSH session warning banner in the Security tab.
- [x] [Frontend] Enhanced Power UI with a 'Power Actions' dropdown (Reboot, Shutdown, Sleep) and safety confirmations.
- [x] [Frontend] Integrated `SecurityPanel` into `VMDetailView.tsx` as a new "Security" tab with 10-second polling.

### Phase 7.1: Security Desync Fix & UI Component Isolation
- [x] [Backend] **Critical fix**: SSH sessions are single-use (`RunCmd` consumes the session). Rewrote `fetchFirewall` to use a single compound shell command for multi-step detection.
- [x] [Backend] **Critical fix**: Rewrote `fetchSELinux` with the same compound-command pattern — detects binaries and runs `sestatus`/`getenforce` in one SSH call.
- [x] [Backend] Firewall detection uses `systemctl is-active firewalld` inside the compound script for robustness.
- [x] [Backend] SELinux detection checks both `command -v` and `/usr/sbin/` full paths for restricted SSH PATH environments.
- [x] [Frontend] Created `CompactPowerActions.tsx` — isolated compact power dropdown for Inventory cards (opens upward, z-index 200).
- [x] [Frontend] Created `FullPowerActions.tsx` — isolated full power dropdown for VM Details page (opens downward, z-index 200).
- [x] [Frontend] Fixed Terminal user-select dropdown in Inventory — now uses inline styles (opens upward, z-index 200) instead of shared `.dropdown-menu` CSS.
- [x] [Frontend] All card dropdowns (Terminal, Power) use self-contained inline positioning to prevent CSS class conflicts.

---

## Final Phase: Polish & Cross-Cutting Concerns

**Purpose**: Integration testing, security hardening, and documentation updates.

- [x] T028 Write a Go integration test for the provisioning engine in `backend/tests/integration/provisioning_test.go` validating: successful provisioning updates status to "provisioned", failed visudo sets status to "failed" with error message, and audit events are logged
- [x] T029 [P] Update `backend/Makefile` to add a `build` target that compiles the binary (ensuring `//go:embed` bundles the provisioning script) and a `test-provision` target
- [x] T030 [P] Update `README.md` with V2 feature documentation covering: zero-touch provisioning, multi-page navigation, network telemetry, and process explorer
- [x] T031 Validate the complete provisioning→telemetry→navigation flow against all quickstart.md scenarios in `specs/002-v2-enterprise-upgrade/quickstart.md`

---

## Explicitly Excluded Tasks

The following task types are **permanently excluded** per Architectural Constraints:

- ❌ No `SudoPromptModal` component task
- ❌ No `sudo_password` API field or endpoint task
- ❌ No sudo credential caching or session task
- ❌ No frontend password input dialog for sudo operations
- ❌ No `NOPASSWD: ALL` sudoers configuration

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
