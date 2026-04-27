# Phase 0 Research: VM Dashboard UI Redesign & Backend Sync

**Branch**: `005-vm-dashboard-ui-sync`  
**Date**: 2026-04-27  

---

## R1: Per-Card Telemetry Fetching Strategy

**Decision**: Fetch stats only for online VMs using staggered requests; skip offline VMs entirely.

**Rationale**: The inventory page fetches the VM list via `GET /api/vms`, then iterates over VMs with `status === "online"` and issues `GET /api/vms/{id}/stats` for each. Requests are staggered (e.g., 50ms delay between each) to avoid flooding the backend. Offline VMs display a "No data" indicator on their cards. This reduces call count proportionally to online VMs and avoids 404 responses for VMs without telemetry agents.

**Alternatives considered**:
- Fetch all VMs stats in parallel: Rejected — causes N+1 API flooding with 50+ VMs
- Intersection Observer / lazy scroll: Rejected — adds complexity, all VMs rendered in one grid
- No per-card metrics: Rejected — reduces information density on cards

---

## R2: OS Indicator on VM Cards

**Decision**: Omit OS-specific indicators; use a generic server icon on all VM cards.

**Rationale**: The backend `VM` model does not include an `os` field. Adding one would violate the "no backend changes" constraint. Telemetry data doesn't reliably expose OS info either. A generic server icon (lucide-react `Server` icon) provides visual consistency without fabricating data.

**Alternatives considered**:
- Derive OS from telemetry: Rejected — unreliable, not all VMs have agents
- Add OS field to backend: Rejected — violates scope constraint
- Hardcode OS based on hostname patterns: Rejected — fragile and misleading

---

## R3: Large VM Grid Rendering Strategy

**Decision**: Render all VMs in a single grid with `React.memo` optimization. No pagination or virtualization.

**Rationale**: With up to 100 VMs, a single grid is manageable with React.memo preventing unnecessary re-renders. The existing grid layout already handles dynamic counts. Adding pagination would break the single-view overview experience. Virtualization adds library dependencies for a scenario that's uncommon.

**Alternatives considered**:
- Pagination (20 per page): Rejected — breaks overview experience
- Virtual scroll: Rejected — adds dependency, overkill for 100 items
- Infinite scroll: Rejected — unnecessary complexity

---

## R4: Glassmorphism Implementation Approach

**Decision**: Extend existing CSS custom properties in `index.css` with glassmorphism utility classes using `backdrop-filter`, semi-transparent backgrounds, and soft borders.

**Rationale**: The project already has a mature CSS design system (`index.css`, 777 lines) with CSS custom properties for colors, backgrounds, borders, and component styles. Glassmorphism effects (backdrop blur, translucent surfaces) can be added as new utility classes and card variants within the existing `@layer components` structure. Tailwind CSS is available for utility overrides but the core design system uses CSS classes.

**Alternatives considered**:
- Tailwind-only glassmorphism: Rejected — inconsistent with existing CSS custom property system
- Third-party glassmorphism library: Rejected — unnecessary dependency

---

## R5: Staggered API Request Pattern

**Decision**: Use a simple `Promise`-based stagger with configurable delay (50ms default) and `Promise.allSettled` for error resilience.

**Rationale**: When fetching stats for N online VMs, requests are dispatched with a 50ms delay between each using a for-loop with `await new Promise(resolve => setTimeout(resolve, 50))`. `Promise.allSettled` is used to ensure individual failures don't block the entire batch. Failed requests result in "No data" on that card.

**Alternatives considered**:
- `Promise.all` without stagger: Rejected — floods backend
- Web Worker for background fetching: Rejected — overkill for this use case
- Service Worker cache: Rejected — adds complexity, data is real-time

---

## R6: ProvisioningBadge Styling Inconsistency

**Decision**: Restyle ProvisioningBadge to use the dark theme design tokens (`prov-badge` CSS classes) instead of the current light-mode Tailwind classes (bg-green-100, bg-red-100, etc.).

**Rationale**: The current ProvisioningBadge component uses light-mode Tailwind utilities (bg-green-100, text-green-800) which visually clash with the dark theme. The `index.css` already defines `.prov-badge--provisioned`, `.prov-badge--pending`, `.prov-badge--failed` classes with proper dark-theme colors. The component should be updated to use these.

**Alternatives considered**:
- Keep light-mode badges: Rejected — visually inconsistent with dark theme

---

## R7: Constitution Compliance — WebSocket for Real-Time Data

**Decision**: The constitution requires WebSocket for "real-time data streams." The inventory page's periodic polling (5s/10s intervals) for VM list updates is acceptable as it's a CRUD list refresh, not a real-time metric stream. Per-card telemetry stats fetched via REST are also acceptable as they're point-in-time snapshots, not continuous streams. The VM detail page's existing telemetry auto-refresh via REST polling should be noted as a potential future WebSocket migration candidate, but is out of scope for this feature.

**Rationale**: Constitution Principle II states "HTTP polling MUST NOT be used for any real-time data stream." The inventory list polling is a periodic CRUD refresh (not a live stream), and per-card stats are on-demand snapshots. The SSH terminal already uses WebSocket correctly. Dashboard stats polling is a CRUD refresh pattern.

**Alternatives considered**:
- Convert all polling to WebSocket: Rejected — violates "no backend changes" constraint; would require new WS endpoints
