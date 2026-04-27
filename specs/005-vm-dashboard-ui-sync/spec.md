# Feature Specification: VM Dashboard UI Redesign & Backend Sync

**Feature Branch**: `005-vm-dashboard-ui-sync`  
**Created**: 2026-04-27  
**Status**: Draft  
**Input**: User description: "Redesign the frontend UI to match an uploaded reference image and align the frontend implementation with the backend enhancements and changes that have already been completed."

---

## Problem Statement

The current VM management dashboard frontend was built incrementally across multiple feature phases (001 through 004). While the backend has been significantly enhanced with telemetry, provisioning, security hardening, and agent-based monitoring capabilities, the frontend has not kept pace visually or functionally. The existing UI uses a functional but basic layout with simple card-based inventory display that does not surface the full depth of available backend data (e.g., OS info, hostname, IP address, CPU/RAM usage per VM card). Additionally, the overall visual design needs a comprehensive overhaul to match a modern, polished, dark-themed admin dashboard reference image.

---

## Clarifications

### Session 2026-04-27

- Q: How should the VM inventory handle large VM counts (100+ VMs)? → A: Render all VMs in the grid with React.memo optimization — no pagination, no virtual scroll.
- Q: How should the OS indicator on VM cards be handled given the backend has no `os` field? → A: Omit OS indicator entirely; use a generic server icon on cards.
- Q: What telemetry fetching strategy should the inventory page use for per-card CPU/RAM metrics? → A: Fetch stats only for online VMs, skip offline; use staggered requests to avoid API flooding.
- Q: Quick power actions in top-right vs per-card? → A: Top-right area is used for global actions only (e.g. Add VM, Refresh). Power actions remain tightly coupled within the VM cards.

---

## Goals

1. **Visual Redesign**: Transform the frontend into a premium, dark-themed admin dashboard that matches the provided reference image — featuring glassmorphism panels, a restructured sidebar, an enhanced top bar, and a card-based VM inventory with richer per-card metadata.

2. **Backend Alignment**: Update all frontend components to consume and display the full range of data now available from the backend API (telemetry metrics, provisioning status, VM metadata, audit events, security data).

3. **Improved Information Density**: Each VM card in the inventory should display actionable metadata at a glance — hostname, IP address, CPU usage, RAM usage, status badges, and per-card action buttons. A generic server icon is used in place of OS-specific indicators since the backend does not provide OS data.

4. **Consistent UX Flow**: Ensure every user-facing action (add/edit/delete VM, power actions, terminal, provisioning retry, user access management) works correctly with the current backend endpoints and response contracts.

5. **State Handling Completeness**: Implement proper loading, error, empty, and success states across all views to provide a professional, resilient user experience.

---

## Non-Goals

- **Backend API Changes**: The backend is treated as stable. No new endpoints or schema modifications are part of this feature. The frontend must adapt to existing backend contracts.
- **New Feature Development**: This feature does not add new backend capabilities (e.g., no new monitoring agents, no new provisioning logic). It surfaces existing capabilities in the UI.
- **Mobile-Native App**: While the dashboard should be responsive, building a dedicated mobile application is out of scope.
- **User Authentication Redesign**: The login flow and auth context remain unchanged; only visual styling of the login page may be updated to match the new design system.
- **Light-mode Toggle**: Light-mode toggle is deferred to a future feature; this redesign establishes the dark-mode-first foundation per constitution, with the toggle as future work.
- **Internationalization (i18n)**: Multi-language support is not in scope.

---

## User Scenarios & Testing

### User Story 1 — Admin Views VM Inventory Dashboard (Priority: P1)

An administrator logs in and navigates to the VM inventory page. They see a modern, dark-themed dashboard with a left sidebar navigation, a top bar with search and quick actions, and a responsive grid of VM cards. Each card shows the VM name, hostname/IP, a generic server icon, live CPU and RAM usage gauges, connectivity status badge (online/offline), provisioning status badge, and per-card action buttons (console, reboot, settings/edit).

**Why this priority**: The VM inventory is the primary view of the application. It is the first thing admins interact with after login, and its redesign is the core visual deliverable of this feature.

**Independent Test**: Can be fully tested by logging in, viewing the VM inventory, and verifying that each VM card displays all required metadata fields and status badges with correct data from the backend. Test with VMs in different states (online, offline, provisioned, pending, failed).

**Acceptance Scenarios**:

1. **Given** an authenticated admin with 5 registered VMs, **When** they navigate to the VMs page, **Then** they see a responsive grid of 5 VM cards, each displaying: name, host, status badge (online/offline), provisioning badge, and action buttons.
2. **Given** a VM that is online with active telemetry data, **When** the inventory loads, **Then** that VM's card shows CPU and RAM usage metrics sourced from the `/api/vms/{id}/stats` endpoint.
3. **Given** a VM with provisioning status "failed", **When** the admin views its card, **Then** a "failed" badge is shown with a retry action available.
4. **Given** no VMs registered, **When** the admin visits the inventory, **Then** an empty state is displayed with a prompt to add the first server.

---

### User Story 2 — Admin Uses Redesigned Navigation & Layout (Priority: P1)

An administrator sees a persistent left sidebar with navigation links (Dashboard, VMs, Audit Log, Key Vault, Settings) and a top bar with the current page title, contextual search, and user profile/actions. The layout features glassmorphism styling, consistent dark theme, and smooth transitions between pages.

**Why this priority**: The layout (sidebar + topbar + content area) is the structural foundation. Every page depends on it, so it must be redesigned first.

**Independent Test**: Navigate between all available routes (Dashboard, VMs, VM Detail, Audit Log, Keys, Settings) and verify the sidebar highlights the active item, the top bar updates the page title, and the search bar appears contextually on the VMs page.

**Acceptance Scenarios**:

1. **Given** the admin is on the Dashboard page, **When** they click "VMs" in the sidebar, **Then** the sidebar highlights the VMs link, the top bar shows "Infrastructure", and the VMs page loads with a smooth transition.
2. **Given** the admin is on the VMs page, **When** they type in the search bar, **Then** the VM grid filters in real time by name, host, or tag.
3. **Given** any page, **When** the admin clicks their profile avatar, **Then** a dropdown appears with user info and a sign-out option.

---

### User Story 3 — Admin Performs VM Actions from Card (Priority: P2)

From the VM inventory, an administrator can perform quick actions on any VM directly from its card: open an SSH terminal, trigger a reboot/shutdown, open settings/edit modal, or delete the VM. Power actions are accessible from compact action buttons on each card, and optionally from a top-right quick-power area.

**Why this priority**: Actions are the primary workflow after viewing VMs. They must be integrated into the new card design and wired to the existing backend API endpoints.

**Independent Test**: From the VM inventory, click each action button (terminal, power, edit, delete) for a VM and verify the correct backend API call is made and the UI responds appropriately (modal opens, confirmation dialog appears, terminal connects, etc.).

**Acceptance Scenarios**:

1. **Given** a VM card, **When** the admin clicks the terminal/console button, **Then** a terminal overlay opens connecting to that VM via WebSocket, or a user-select dropdown appears if authorized_users exist.
2. **Given** a VM card, **When** the admin clicks the reboot action, **Then** a confirmation dialog appears, and upon confirmation, a POST to `/api/vms/{id}/power` with `{"action": "reboot"}` is sent.
3. **Given** a VM card, **When** the admin clicks edit/settings, **Then** the edit modal opens pre-populated with the VM's current data (name, host, username, auth type, tags, key_id).
4. **Given** a VM card, **When** the admin clicks delete, **Then** a destructive confirmation dialog appears, and upon confirmation, a DELETE to `/api/vms/{id}` is sent, and the VM is removed from the grid.

---

### User Story 4 — Admin Views Dashboard Overview (Priority: P2)

An administrator visits the Dashboard overview page and sees aggregate statistics (total VMs, online count, offline count, provisioned count, pending count, failed provisioning alerts, audit event count) displayed as hero metric cards with the new visual design. If there are failed provisioning VMs, an alert banner is shown with a link to the inventory.

**Why this priority**: The dashboard overview is the landing page and provides a high-level health snapshot. Its data is already available from the `/api/dashboard/stats` endpoint.

**Independent Test**: Navigate to the dashboard and verify all 6 hero cards display correct counts matching the database state. Verify the alert banner appears when failed provisioning count > 0.

**Acceptance Scenarios**:

1. **Given** 10 VMs (7 online, 3 offline, 8 provisioned, 1 pending, 1 failed), **When** the admin loads the dashboard, **Then** each hero card shows the correct count with the new visual styling.
2. **Given** 1 VM with failed provisioning, **When** the dashboard loads, **Then** a warning banner is displayed with a link to the VMs page.

---

### User Story 5 — Admin Views VM Detail Page with Telemetry (Priority: P3)

An administrator clicks a VM from the inventory (or navigates to `/vms/{id}`) and sees a detailed view with: live CPU/RAM/Disk gauges, network metrics, process table, service list, log viewer, and security panel — all sourced from the backend telemetry and monitoring APIs.

**Why this priority**: The detail view is a deeper drill-down. It's important but secondary to the inventory view redesign.

**Independent Test**: Navigate to a VM detail page for an online VM with active telemetry data. Verify CPU, RAM, disk, and network metrics are displayed and auto-refresh. Verify process table, service list, and log viewer populate correctly.

**Acceptance Scenarios**:

1. **Given** an online VM with telemetry data, **When** the admin navigates to its detail page, **Then** CPU/RAM/disk usage gauges display current values from `/api/vms/{id}/stats`.
2. **Given** a VM that is offline, **When** the admin views its detail page, **Then** a graceful "No data" message is displayed instead of empty/broken gauges.

---

### Edge Cases

- What happens when a VM's telemetry data has not been received yet? (Both "awaiting data" and offline "no data" states share the same "—" indicator on the metric gauges to maintain UX simplicity).
- How does the card handle extremely long VM names or hostnames? (Truncate with ellipsis)
- What happens if the backend returns an empty tags array? (Display "No tags" placeholder)
- How does the grid behave with 100+ VMs? → Render all VMs with React.memo optimization; no pagination or virtual scroll needed.
- What happens if the dashboard stats API fails? (Show skeleton/error state, retry after interval)
- What happens if the user loses their auth token mid-session? (Covered by existing infrastructure: AuthContext redirects to login)
- How does the sidebar behave on narrow screens? (Collapse to icon-only or hamburger menu)

---

## Requirements

### Functional Requirements

- **FR-001**: The application MUST render a persistent left sidebar with navigation links to Dashboard, VMs, Audit Log, Key Vault, and Settings pages.
- **FR-002**: The application MUST render a top navigation bar showing the current page title, contextual search (on VMs page), and a user profile menu with sign-out capability.
- **FR-003**: The VM inventory page MUST display VMs in a responsive card grid (1 column mobile, 2 columns tablet, 3 columns desktop).
- **FR-004**: Each VM card MUST display: VM name, hostname/IP, connectivity status badge (online/offline), provisioning status badge, and action buttons.
- **FR-005**: Each VM card SHOULD display CPU usage and RAM usage when telemetry data is available. Telemetry is fetched only for VMs with `status === "online"` using staggered requests (not all at once). Offline VMs display a "No data" indicator. A generic server icon is used on all cards (no OS-specific indicator, as the backend does not provide OS data).
- **FR-006**: VM card action buttons MUST include: console/terminal, power actions (reboot, shutdown, sleep), edit/settings, and delete.
- **FR-007**: The terminal button MUST present a user selection dropdown when `authorized_users` array is non-empty, falling back to the management username.
- **FR-008**: The Dashboard overview page MUST display hero metric cards for: Total VMs, Online, Offline, Provisioned, Pending, and Audit Events — sourced from the `/api/dashboard/stats` endpoint.
- **FR-009**: All API calls MUST use the existing axios-based API service with Bearer token authentication from the AuthContext.
- **FR-010**: The application MUST handle loading states (skeletons), error states (user-friendly messages with retry), and empty states (prompts to add data) for all data-driven views.
- **FR-011**: The Add/Edit VM modal MUST submit requests matching the backend `CreateVMRequest` and `UpdateVMRequest` contracts (name, host, management_username, auth_type, credential, key_id, tags).
- **FR-012**: The delete VM flow MUST display a destructive confirmation dialog before sending the DELETE request.
- **FR-013**: Power actions (reboot, shutdown, sleep) MUST send `POST /api/vms/{id}/power` with the `{"action": "..."}` payload and display appropriate confirmation.
- **FR-014**: Provisioning status badges MUST render with visual distinction for states: "provisioned", "pending", "not_provisioned", "failed" — with a retry action available for "failed" status.
- **FR-015**: The search functionality on the VMs page MUST filter the local VM list by name, host, or tags.
- **FR-016**: The VM inventory MUST poll for status updates at 5-second intervals when VMs have "pending" provisioning, and 10-second intervals otherwise.
- **FR-017**: The VM detail page MUST display telemetry data (CPU, RAM, disk, network) sourced from `/api/vms/{id}/stats` with auto-refresh using HTTP polling (per accepted deviation).

### Key Entities

- **VM**: Core entity — `id`, `name`, `host`, `management_username`, `auth_type`, `status` (online/offline), `provisioning_status`, `tags[]`, `authorized_users[]`, `key_id`, `created_at`
- **DashboardStats**: Aggregate counts — `total_vms`, `online_vms`, `offline_vms`, `provisioned`, `pending`, `failed`, `audit_events`
- **VMStats**: Telemetry data — `cpu` (%), `ram` (%), `disk` (%), `network` (rx/tx bytes/sec, active connections), `processes[]`, `timestamp`
- **AuditLog**: Event record — `id`, `event_type`, `details`, `timestamp`
- **AgentToken**: OTT for enrollment — `id`, `machine_id`, `created_at`, `expires_at`, `used_at`

---

## UI Requirements

### Visual Design Direction

The redesigned UI must achieve a premium, dark-themed admin dashboard aesthetic matching the provided reference image. The following visual characteristics are required:

- **Color Palette**: Deep dark backgrounds (slate-950/900 range), accent colors for status indicators (emerald for online, amber for pending, red for failed/offline, blue for primary actions, violet for provisioned).
- **Glassmorphism Panels**: Semi-translucent card and panel surfaces with subtle backdrop blur, soft borders, and layered depth through shadows and glow effects.
- **Typography**: Modern sans-serif font (Inter or similar from Google Fonts), with a clear hierarchy — large bold titles, medium section headers, small metadata labels.
- **Card Design**: Rounded corners (xl to 2xl radius), consistent internal padding, hover elevation effects with subtle glow, and a clean information hierarchy within each card.
- **Sidebar**: Full-height left sidebar with brand/logo area at top, vertically stacked navigation items with icon + label, active state indicator (highlight or accent bar), and a system status footer.
- **Top Bar**: Horizontal bar spanning the content area with page title, search input (contextual), and right-aligned action icons (server count, key vault shortcut, user avatar with dropdown).
- **Status Badges**: Pill-shaped badges with colored dot + text, distinct per status (online = emerald, offline = slate, provisioned = violet, pending = amber, failed = red).
- **Action Buttons**: Compact icon buttons with hover states, grouped in the card footer area. Primary action (terminal) should be visually emphasized.
- **Animations**: Smooth page transitions (enter animations), hover micro-animations on cards and buttons, and skeleton loading states during data fetch.
- **Spacing & Rhythm**: Consistent spacing scale, adequate whitespace between sections, and a clear visual rhythm that guides the eye from sidebar → top bar → page title → content grid.

### Responsive Behavior

- **Desktop (≥1280px)**: Full sidebar expanded, 3-column VM card grid, full top bar.
- **Tablet (768px–1279px)**: Sidebar collapsed to icon-only or hidden with hamburger toggle, 2-column grid.
- **Mobile (<768px)**: Sidebar hidden behind hamburger menu, 1-column grid, stacked layout for card content.

---

## Non-Functional Requirements

- **NFR-001**: Page initial load (time to interactive) MUST be under 3 seconds on a standard broadband connection.
- **NFR-002**: UI interactions (button clicks, modal opens, page transitions) MUST feel instant with no perceptible lag (< 100ms response).
- **NFR-003**: The dashboard MUST remain functional and responsive with up to 100 VMs rendered simultaneously in a single grid (no pagination), using component memoization to prevent unnecessary re-renders.
- **NFR-004**: All styles MUST follow the existing design token / CSS variable system already in the project's `index.css`.
- **NFR-005**: The application MUST gracefully degrade when backend APIs are unreachable — showing cached data or informative error states, never blank screens.

---

## Backend/Frontend Integration Requirements

### Existing Backend API Contracts (No Changes Required)

The frontend must integrate with these existing backend endpoints:

| Endpoint | Method | Request Body | Response | Frontend Usage |
|----------|--------|-------------|----------|----------------|
| `/api/vms` | GET | — | `VM[]` | Inventory grid |
| `/api/vms` | POST | `CreateVMRequest` | `VM` + provisioning_status | Add VM modal |
| `/api/vms/{id}` | GET | — | `VM` (with credential) | VM detail page |
| `/api/vms/{id}` | PUT | `UpdateVMRequest` | 204 No Content | Edit VM modal |
| `/api/vms/{id}` | DELETE | — | 204 No Content | Delete VM flow |
| `/api/vms/{id}/stats` | GET | — | `VMStats` | Card metrics + detail page |
| `/api/vms/{id}/power` | POST | `{"action": "..."}` | `{"status":"success","message":"..."}` | Power action buttons |
| `/api/vms/{id}/users` | GET | — | `string[]` | User management in detail |
| `/api/vms/{id}/users` | POST | `{"username":"..."}` | 201 Created | Add user |
| `/api/vms/{id}/users/{username}` | DELETE | — | 204 No Content | Revoke user |
| `/api/vms/{id}/provision` | POST | — | — | Retry provisioning |
| `/api/dashboard/stats` | GET | — | `DashboardStats` | Dashboard hero cards |
| `/api/audit-logs` | GET | `?limit=&offset=` | `{logs: AuditLog[], total: int}` | Audit log page |
| `/api/login` | POST | `{username, password}` | `{token}` | Login page |
| `/api/keys` | GET/POST/DELETE | varies | varies | Key vault page |

### Integration Mapping

- **VM Card → `/api/vms` (list)** + **`/api/vms/{id}/stats` (per-card metrics)**: The inventory page fetches the VM list, then fetches stats only for online VMs using staggered requests to prevent API flooding. Offline VMs show a "No data" indicator on their cards.
- **Dashboard → `/api/dashboard/stats`**: Single endpoint provides all hero card counts.
- **VM Detail → `/api/vms/{id}` + `/api/vms/{id}/stats`**: Detail page combines VM metadata with live telemetry.
- **Actions → Power/Provisioning/User endpoints**: All action buttons map to existing POST/DELETE endpoints.
- **Auth → `/api/login` + Bearer token in localStorage**: Existing AuthContext and axios interceptor handle authentication.

### Data Binding & Status Mapping

- `vm.status` values: `"online"` | `"offline"` → maps to connectivity badge colors
- `vm.provisioning_status` values: `"provisioned"` | `"pending"` | `"not_provisioned"` | `"failed"` → maps to provisioning badge variants
- `vmStats.cpu`, `vmStats.ram`, `vmStats.disk` → percentage values (0-100) for gauge/progress displays
- `vm.tags[]` → rendered as small pill tags on cards
- `vm.authorized_users[]` → determines terminal user-select dropdown behavior

---

## Assumptions

- The backend API is stable and fully functional. All endpoints listed above are operational and return the documented response shapes.
- The reference image will be provided again during implementation for pixel-level guidance.
- The existing Vite + React + TypeScript + Tailwind CSS tech stack will be retained.
- The existing `index.css` design token system (CSS custom properties) will be extended, not replaced.
- Telemetry data may not be available for all VMs (e.g., agent not deployed yet). The UI must handle missing telemetry gracefully.
- The existing component library (lucide-react icons, axios) will continue to be used.
- WebSocket-based SSH terminal functionality (TerminalView component) is stable and requires only visual restyling, not functional changes.

---

## Constraints & Accepted Deviations

- **Constraint: No Backend Changes**: No new backend API endpoints may be created as part of this feature. The frontend must adapt to existing backend contracts.
- **Accepted Deviation: HTTP Polling for Telemetry**: Principle 5 states that WebSocket is preferred for live metrics. Since the backend does not currently provide WebSocket telemetry endpoints and we are constrained to making no backend changes, fetching live telemetry metrics (CPU/RAM) via REST `GET /api/vms/{id}/stats` polling is documented as an accepted interim solution.
- **Constraint**: The frontend must remain a single-page application (SPA) deployed via Vite.
- **Constraint**: All changes must be backward-compatible with the existing backend version.
- **Constraint**: The Tailwind CSS configuration and PostCSS pipeline are already in place and should be leveraged.

---

## Dependencies

- Existing backend server running with all API endpoints operational.
- Backend database schema (vms, provisioning_records, agent_tokens, agent_certificates, audit_log_entries, vm_access) is stable.
- Reference UI image (to be provided during implementation phase).
- Google Fonts CDN for typography (Inter).

---

## Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Telemetry endpoint returns 404 for VMs without agents | High | Medium | Implement graceful fallback — show "No data" on card instead of error |
| Per-card stats fetching creates N+1 API call pattern with many VMs | Medium | Medium | Fetch stats only for online VMs with staggered requests; offline VMs show "No data" — reduces call count proportionally |
| Reference image interpretation differences during implementation | Medium | Medium | Treat the image as directional, not pixel-perfect; focus on design system consistency |
| Large VM counts (100+) cause performance issues with grid re-render | Low | Medium | Use React.memo for memoization; no virtualization needed per clarification. |
| WebSocket terminal connection failures during SSH | Low | High | Existing error handling in TerminalView should be preserved and enhanced visually |

---

## Success Criteria

### Measurable Outcomes

- **SC-001**: The VM inventory page displays all registered VMs in a card grid with name, host, status badge, provisioning badge, and action buttons — matching the reference image's layout structure, visual hierarchy, and component arrangement without introducing new layout paradigms.
- **SC-002**: An administrator can complete the full VM lifecycle (view → add → edit → terminal → power action → delete) in under 5 minutes without encountering any UI errors.
- **SC-003**: All 6 dashboard hero cards display correct, live-updating counts that match the database state at any point in time.
- **SC-004**: The UI remains responsive and interactive with up to 100 VMs in the inventory, with no perceptible lag when scrolling, filtering, or performing actions.
- **SC-005**: Every data-driven view (inventory, dashboard, detail, audit log, key vault) properly handles loading, error, and empty states without showing blank screens or unformatted data.
- **SC-006**: The redesigned UI passes a visual consistency check — all pages share the same color palette, typography, spacing system, and component styling.
- **SC-007**: 100% of existing backend endpoints are correctly consumed by the frontend with no broken API integrations after the redesign.

---

## Open Questions

None. All open questions resolved.
