# Tasks: VM Dashboard UI Redesign & Backend Sync

**Input**: Design documents from `/specs/005-vm-dashboard-ui-sync/`  
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/frontend-ui.md  
**Tests**: Not explicitly requested — test tasks omitted, manual verification at checkpoints.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Design System Foundation)

**Purpose**: Extend the CSS design system with glassmorphism tokens, new component classes, and metric gauge styles. All subsequent phases depend on this.

- [ ] T001 Add glassmorphism CSS custom properties (`--glass-bg`, `--glass-border`, `--glass-blur`, `--glow-blue`, `--glow-emerald`, `--glow-red`) to `:root` in `frontend/src/index.css`
- [ ] T002 Add `.glass-panel` and `.glass-card` component classes with `backdrop-filter: blur()`, semi-transparent backgrounds, and soft borders in `frontend/src/index.css`
- [ ] T003 Add `.vm-card`, `.vm-card-header`, `.vm-card-body`, `.vm-card-actions` component classes with hover glow effects in `frontend/src/index.css`
- [ ] T004 Add `.metric-gauge`, `.metric-gauge-fill`, `.metric-gauge-label` classes for compact inline CPU/RAM progress bars in `frontend/src/index.css`
- [ ] T005 Add responsive sidebar CSS — media queries for sidebar collapse at 1279px (icon-only) and hide at 768px (hamburger) in `frontend/src/index.css`
- [ ] T006 Verify all existing design tokens (`.premium-card`, `.hero-card`, `.status-badge`, `.prov-badge`, `.topbar`, `.sidebar`) remain functional after new additions in `frontend/src/index.css`

---

## Phase 2: Foundational (Layout & Navigation — Blocking Prerequisites)

**Purpose**: Restructure the app shell (sidebar + topbar + content area). Every page and user story depends on this layout being correct.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [ ] T007 Restyle `Sidebar.tsx` — enhanced brand area with gradient icon, improved nav link hover/active states using design tokens, system status footer with glow dot in `frontend/src/components/Sidebar.tsx`
- [ ] T008 Restyle `Layout.tsx` TopBar — page title rendering, contextual search input (visible only on `/vms` route), user profile avatar with dropdown menu, server count badge icon, key vault shortcut icon in `frontend/src/components/Layout.tsx`
- [ ] T009 Add responsive sidebar behavior to `Layout.tsx` and `Sidebar.tsx` — tablet (≤1279px): icon-only collapsed sidebar; mobile (≤768px): sidebar hidden behind hamburger toggle button in `frontend/src/components/Layout.tsx` and `frontend/src/components/Sidebar.tsx`
- [ ] T010 Verify page transitions — ensure `page-enter` CSS animation works correctly with new layout structure across all routes in `frontend/src/components/Layout.tsx`

**Checkpoint**: Layout shell fully functional — sidebar navigation, topbar, and content area render correctly at all breakpoints. All existing pages still accessible.

---

## Phase 3: User Story 1 — Admin Views VM Inventory Dashboard (Priority: P1) 🎯 MVP

**Goal**: Rebuild the VM inventory page with a responsive card grid, per-card telemetry gauges for online VMs, and proper status badges. This is the primary visual deliverable.

**Independent Test**: Log in, navigate to VMs page, verify: responsive card grid renders all VMs with name, host, status badge, provisioning badge, action buttons. Verify CPU/RAM gauges appear on online VM cards with telemetry data. Verify empty state renders when no VMs exist. Verify search filters by name, host, tags.

### New Components for User Story 1

- [ ] T011 [P] [US1] Create `StatusBadge.tsx` — reusable online/offline pill badge with colored dot + text, using `.status-badge--online` and `.status-badge--offline` CSS classes. Props: `{ status: "online" | "offline" }` in `frontend/src/components/StatusBadge.tsx`
- [ ] T012 [P] [US1] Create `MetricGauge.tsx` — compact inline progress bar with percentage label for CPU/RAM display on cards. Props: `{ label: string; value: number; color: "blue" | "violet" | "amber" }`. Shows "—" when value is undefined. Uses `.metric-gauge` CSS classes in `frontend/src/components/MetricGauge.tsx`
- [ ] T013 [P] [US1] Restyle `ProvisioningBadge.tsx` — replace light-mode Tailwind utilities (bg-green-100, bg-red-100, etc.) with dark-theme `.prov-badge--provisioned`, `.prov-badge--pending`, `.prov-badge--failed` CSS classes. Keep existing props and logic unchanged in `frontend/src/components/ProvisioningBadge.tsx`
- [ ] T014 [P] [US1] Extract `DeleteConfirmDialog.tsx` — move the DeleteConfirmDialog component from `Inventory.tsx` into its own file. Props: `{ vm: VM; onConfirm: () => void; onCancel: () => void }` in `frontend/src/components/DeleteConfirmDialog.tsx`

### Core VMCard Component

- [ ] T015 [US1] Create `VMCard.tsx` — new extracted VM card component with: generic server icon, VM name (truncated), hostname/IP (mono, truncated), `StatusBadge`, `ProvisioningBadge` (with retry for failed), CPU/RAM `MetricGauge` (when stats prop provided), tags pills (show subtle "No tags" placeholder if empty), action buttons (terminal, power, edit, delete). Wrap with `React.memo` and custom comparator on `vm.id`, `vm.status`, `vm.provisioning_status`, `stats?.cpu`, `stats?.ram`. Props per contract: `{ vm: VM; stats?: VMStats; onTerminal; onEdit; onDelete; onPowerAction; onRetryProvisioning }` in `frontend/src/components/VMCard.tsx`

### Inventory Page Rebuild

- [ ] T016 [US1] Rebuild `Inventory.tsx` — replace inline card rendering with `VMCard` component grid (1-col mobile, 2-col tablet, 3-col desktop). Add `vmStats: Map<string, VMStats>` state. Import and use `VMCard`, `DeleteConfirmDialog`, `AddVmModal`, `TerminalView`, `CompactPowerActions`. Preserve all existing functionality (add, edit, delete, terminal, power, provisioning retry) in `frontend/src/pages/Inventory.tsx`
- [ ] T017 [US1] Implement staggered stats fetching in `Inventory.tsx` — after VM list loads via `GET /api/vms`, iterate online VMs (`status === "online"`) and fetch `GET /api/vms/{id}/stats` with 50ms stagger delay. Use `Promise.allSettled` for error resilience. Store results in `vmStats` Map state. Handle 404/errors gracefully (skip card, show "No data") in `frontend/src/pages/Inventory.tsx`
- [ ] T018 [US1] Implement optimized HTTP polling in `Inventory.tsx` — 5s interval when any VM has `provisioning_status === "pending"`, 10s otherwise. Re-fetch stats via HTTP polling for online VMs on each poll cycle. Use AbortController to cancel in-flight requests when new poll starts in `frontend/src/pages/Inventory.tsx`
- [ ] T019 [US1] Verify search filtering in `Inventory.tsx` — ensure VM grid filters by name, host, or tags using `searchTerm` from Layout outlet context in `frontend/src/pages/Inventory.tsx`
- [ ] T020 [US1] Verify empty state in `Inventory.tsx` — styled empty state with server icon and "Add server" prompt when `filteredVms.length === 0` in `frontend/src/pages/Inventory.tsx`
- [ ] T021 [US1] Verify terminal user-select dropdown — ensure dropdown appears when `authorized_users.length > 0` with management user listed first plus all authorized users. Verify WebSocket terminal opens correctly in `frontend/src/pages/Inventory.tsx`

**Checkpoint**: VM inventory page fully functional with new card design, per-card telemetry for online VMs, staggered fetching, optimized polling, search, empty state, and all action buttons working. This is the MVP.

---

## Phase 4: User Story 2 — Admin Uses Redesigned Navigation & Layout (Priority: P1)

**Goal**: Verify and polish the layout shell built in Phase 2 to meet all acceptance criteria — sidebar active state, topbar page title, contextual search, profile dropdown.

**Independent Test**: Navigate between all routes (Dashboard, VMs, VM Detail, Audit Log, Keys, Settings) and verify: sidebar highlights active item, topbar shows correct page title, search bar appears only on VMs page, profile avatar dropdown works with sign-out.

**Note**: Most layout work was done in Phase 2 (Foundational). This phase handles polish and acceptance validation.

- [ ] T022 [US2] Verify sidebar active state highlighting — ensure correct nav item is highlighted for all routes including nested routes (`/vms/:id`) in `frontend/src/components/Sidebar.tsx`
- [ ] T023 [US2] Verify topbar page title mapping — ensure `getPageTitle()` returns correct title for all routes: Dashboard, Infrastructure, VM Details, Audit Log, Key Vault, Settings in `frontend/src/components/Layout.tsx`
- [ ] T024 [US2] Verify contextual search visibility — search input appears only on `/vms` route and is hidden on all other routes in `frontend/src/components/Layout.tsx`
- [ ] T025 [US2] Verify profile dropdown — avatar click opens dropdown with "Admin User" label and "Sign Out" button. Sign out calls `logout()` and navigates to `/login` in `frontend/src/components/Layout.tsx`
- [ ] T026 [US2] Verify responsive layout at all breakpoints — desktop (≥1280px): full sidebar + 3-col grid; tablet (768–1279px): collapsed sidebar + 2-col grid; mobile (<768px): hidden sidebar + hamburger + 1-col grid in `frontend/src/components/Layout.tsx` and `frontend/src/components/Sidebar.tsx`

**Checkpoint**: Navigation and layout fully polished — active states, page titles, search, profile dropdown, and responsive behavior all verified.

---

## Phase 5: User Story 3 — Admin Performs VM Actions from Card (Priority: P2)

**Goal**: Ensure all VM action buttons (terminal, power, edit, delete) work correctly with the new card design and existing backend endpoints.

**Independent Test**: From inventory, click each action button for a VM and verify: terminal opens (with user-select if authorized_users), power action sends correct POST, edit modal opens pre-populated, delete shows confirmation then removes VM.

- [ ] T027 [P] [US3] Restyle `CompactPowerActions.tsx` — update dropdown styling to use CSS custom properties instead of inline styles, match new glass-panel design, verify reboot/sleep/shutdown buttons send correct `POST /api/vms/{id}/power` payload in `frontend/src/components/CompactPowerActions.tsx`
- [ ] T028 [P] [US3] Restyle `AddVmModal.tsx` — apply dark theme styling with glassmorphism backdrop, ensure form fields match `CreateVMRequest`/`UpdateVMRequest` contracts (name, host, management_username, auth_type, credential, key_id, tags), verify pre-population when editing in `frontend/src/components/AddVmModal.tsx`
- [ ] T029 [P] [US3] Restyle `TerminalView.tsx` — apply dark theme to terminal wrapper/chrome only, do NOT modify WebSocket connection logic or xterm configuration. Verify terminal opens and connects correctly in `frontend/src/components/TerminalView.tsx`
- [ ] T030 [US3] Restyle `ConnectionOverlay.tsx` — update connection status overlay to use new design tokens in `frontend/src/components/ConnectionOverlay.tsx`
- [ ] T031 [US3] Verify all VM actions end-to-end — from a VM card: terminal opens, power actions trigger with confirmation, edit modal saves, delete confirmation removes VM from grid. All API calls match documented contracts in `frontend/src/pages/Inventory.tsx`

**Checkpoint**: All VM actions (terminal, power, edit, delete) work correctly from the redesigned card layout with proper backend API integration.

---

## Phase 6: User Story 4 — Admin Views Dashboard Overview (Priority: P2)

**Goal**: Restyle the dashboard overview with glassmorphism hero cards showing correct aggregate metrics from the backend.

**Independent Test**: Navigate to dashboard, verify 6 hero cards (Total VMs, Online, Offline, Provisioned, Pending, Audit Events) show correct counts from `/api/dashboard/stats`. Verify failure alert banner appears when failed > 0.

- [ ] T032 [US4] Restyle `DashboardView.tsx` — apply glassmorphism styling to hero metric cards using `.glass-card` class, enhance skeleton loading state, restyle failure alert banner with link to VMs page in `frontend/src/pages/DashboardView.tsx`
- [ ] T033 [US4] Verify dashboard data binding — all 6 hero cards display correct counts from `GET /api/dashboard/stats` response (total_vms, online_vms, offline_vms, provisioned, pending, audit_events). Verify auto-refresh at 15s interval in `frontend/src/pages/DashboardView.tsx`
- [ ] T034 [US4] Verify dashboard alert banner — banner appears when `stats.failed > 0` with warning text and styled link to `/vms` page in `frontend/src/pages/DashboardView.tsx`

**Checkpoint**: Dashboard overview displays correct live-updating metrics with glassmorphism styling. Alert banner works for failed provisioning.

---

## Phase 7: User Story 5 — Admin Views VM Detail Page with Telemetry (Priority: P3)

**Goal**: Restyle the VM detail page with new design tokens, ensuring telemetry data (CPU/RAM/Disk/Network), process table, service table, log viewer, and security panel render correctly.

**Independent Test**: Navigate to a VM detail page for an online VM. Verify CPU/RAM/disk gauges show current values from `/api/vms/{id}/stats`. Verify process table, service table, log viewer populate. Verify offline VMs show "No telemetry data" message.

- [ ] T035 [P] [US5] Restyle `VMDetailView.tsx` — apply new design tokens to detail header (back button, VM name, host, status badge), overview fields grid, tab bar, and metric gauges. Verify telemetry data auto-refresh from `GET /api/vms/{id}/stats` in `frontend/src/pages/VMDetailView.tsx`
- [ ] T036 [P] [US5] Restyle `ProcessTable.tsx` — apply `.data-table` dark theme styling, ensure process list renders correctly from `VMStats.processes[]` in `frontend/src/components/ProcessTable.tsx`
- [ ] T037 [P] [US5] Restyle `ServiceTable.tsx` — apply `.data-table` dark theme styling in `frontend/src/components/ServiceTable.tsx`
- [ ] T038 [P] [US5] Restyle `LogViewer.tsx` — apply dark theme styling to log display area in `frontend/src/components/LogViewer.tsx`
- [ ] T039 [P] [US5] Restyle `NetworkChart.tsx` — apply dark theme to chart using recharts theming, ensure rx/tx bytes data renders correctly in `frontend/src/components/NetworkChart.tsx`
- [ ] T040 [P] [US5] Restyle `SecurityPanel.tsx` — apply dark theme to security information panel in `frontend/src/components/SecurityPanel.tsx`
- [ ] T041 [P] [US5] Restyle `VMDetails.tsx` side panel — apply dark theme styling to the VM details side panel component in `frontend/src/components/VMDetails.tsx`
- [ ] T042 [P] [US5] Restyle `FullPowerActions.tsx` — apply dark theme to full power actions component in `frontend/src/components/FullPowerActions.tsx`
- [ ] T043 [US5] Verify offline VM graceful handling — when navigating to detail page of an offline VM without telemetry, show "No telemetry data available" message instead of empty/broken gauges in `frontend/src/pages/VMDetailView.tsx`

**Checkpoint**: VM detail page fully restyled with working telemetry, process/service tables, log viewer, network chart, and security panel. Offline VMs handled gracefully.

---

## Phase 8: Secondary Pages Restyle

**Purpose**: Apply the new design system to remaining pages (Audit Log, Key Vault, Settings, Login) for full visual consistency.

- [ ] T044 [P] Restyle `AuditLogPage.tsx` — apply dark theme to audit table, event badges, pagination controls. Verify data from `GET /api/audit-logs` in `frontend/src/pages/AuditLogPage.tsx`
- [ ] T045 [P] Restyle `KeysManagement.tsx` — apply dark theme to SSH key cards, add/delete key forms. Verify CRUD operations with `/api/keys` endpoints in `frontend/src/pages/KeysManagement.tsx`
- [ ] T046 [P] Restyle `SettingsPage.tsx` — apply dark theme to settings cards using new design tokens in `frontend/src/pages/SettingsPage.tsx`
- [ ] T047 [P] Restyle `Login.tsx` — apply dark theme to login form (centered card, dark background, styled inputs). Preserve auth functionality (POST `/api/login`, token storage) in `frontend/src/pages/Login.tsx`

**Checkpoint**: All pages share the same color palette, typography, spacing system, and component styling (SC-006).

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: Final integration verification, responsive testing, and visual consistency validation.

- [ ] T048 Run TypeScript type-check — execute `npx tsc --noEmit` from `frontend/` directory and fix any type errors
- [ ] T049 Verify all API integrations — sweep through every endpoint in the API contract table (spec) and confirm the frontend correctly consumes each with no broken calls or mismatched payloads
- [ ] T050 Full VM lifecycle test — execute complete flow: add VM → view in grid → view details → open terminal → execute power action → edit VM → delete VM. Verify each step works end-to-end
- [ ] T051 State handling audit — verify loading skeletons, error states (backend down), and empty states (no VMs, no audit logs, no keys) across all pages. Verify no blank screens when APIs fail
- [ ] T052 Responsive testing across all pages — extends T026 checks by testing all secondary pages and modals at desktop (≥1280px), tablet (768–1279px), and mobile (<768px) breakpoints. Verify sidebar collapse/hide, grid column changes, and stacked layouts
- [ ] T053 Visual consistency check — verify all pages share same color palette, typography, spacing system, and component styling per the reference image design direction
- [ ] T054 Performance spot check — verify UI remains responsive with up to 100 VMs in inventory. Verify no perceptible lag on card hover, search filtering, modal opens, and page transitions

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies — start immediately
- **Phase 2 (Foundational)**: Depends on Phase 1 — BLOCKS all user stories
- **Phase 3 (US1 — Inventory)**: Depends on Phase 2
- **Phase 4 (US2 — Layout Polish)**: Depends on Phase 2 (can run parallel with Phase 3)
- **Phase 5 (US3 — VM Actions)**: Depends on Phase 3 (needs VMCard)
- **Phase 6 (US4 — Dashboard)**: Depends on Phase 2 (can run parallel with Phase 3-5)
- **Phase 7 (US5 — VM Detail)**: Depends on Phase 2 (can run parallel with Phase 3-6)
- **Phase 8 (Secondary Pages)**: Depends on Phase 1 (can run parallel with Phase 3-7)
- **Phase 9 (Polish)**: Depends on ALL previous phases

### User Story Dependencies

- **US1 (P1 — Inventory)**: Depends on Foundational. No dependencies on other stories.
- **US2 (P1 — Layout)**: Depends on Foundational. Built in Phase 2, polished in Phase 4. Independent.
- **US3 (P2 — Actions)**: Depends on US1 (uses VMCard from Phase 3). Action components restyled independently.
- **US4 (P2 — Dashboard)**: Depends on Foundational only. Fully independent from US1/US2/US3.
- **US5 (P3 — Detail)**: Depends on Foundational only. Fully independent from US1/US2/US3/US4.

### Within Each User Story

- New components before page that uses them
- CSS classes before components that reference them
- Core rendering before API integration
- API integration before polling/optimization

### Parallel Opportunities

- **Phase 1**: T001-T005 can all run in parallel (different CSS sections)
- **Phase 3**: T011-T014 can all run in parallel (different component files)
- **Phase 5**: T027-T029 can all run in parallel (different component files)
- **Phase 7**: T035-T042 can all run in parallel (different component files)
- **Phase 8**: T044-T047 can all run in parallel (different page files)
- **Cross-phase**: US4 (Phase 6) can run parallel with US1 (Phase 3-4)
- **Cross-phase**: US5 (Phase 7) can run parallel with US3 (Phase 5)

---

## Parallel Example: User Story 1

```bash
# Launch all new components in parallel (different files):
Task T011: "Create StatusBadge.tsx"
Task T012: "Create MetricGauge.tsx"
Task T013: "Restyle ProvisioningBadge.tsx"
Task T014: "Extract DeleteConfirmDialog.tsx"

# Then sequential (VMCard depends on components above):
Task T015: "Create VMCard.tsx"

# Then sequential (Inventory depends on VMCard):
Task T016: "Rebuild Inventory.tsx grid"
Task T017: "Implement staggered stats fetching"
Task T018: "Implement optimized polling"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Design System Foundation (T001-T006)
2. Complete Phase 2: Layout & Navigation (T007-T010)
3. Complete Phase 3: User Story 1 — Inventory (T011-T021)
4. **STOP and VALIDATE**: VM inventory page works with new card design, telemetry, search, actions
5. Deploy/demo if ready — this is the primary deliverable

### Incremental Delivery

1. Setup + Foundational → Layout shell ready
2. Add US1 (Inventory) → Test independently → **MVP!**
3. Add US2 (Layout polish) → Polish navigation → Demo
4. Add US3 (VM Actions) → Verify all actions → Demo
5. Add US4 (Dashboard) → Dashboard restyled → Demo
6. Add US5 (VM Detail) → Detail page restyled → Demo
7. Secondary Pages + Polish → Full visual consistency → Release

### Parallel Team Strategy

With multiple developers after Phase 2 completes:
- **Developer A**: US1 (Phase 3) → US3 (Phase 5, needs VMCard from US1)
- **Developer B**: US4 (Phase 6) + US2 (Phase 4)
- **Developer C**: US5 (Phase 7) + Phase 8 (Secondary Pages)
- **All**: Phase 9 (Polish) together

---

## Notes

- [P] tasks = different files, no dependencies on other tasks in same phase
- [Story] label maps task to specific user story for traceability
- No backend changes — all work is in `frontend/src/`
- Reference image should be provided during implementation for visual guidance
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Design tokens in `index.css` must be added BEFORE any component work
