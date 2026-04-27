# Quickstart: VM Dashboard UI Redesign & Backend Sync

**Branch**: `005-vm-dashboard-ui-sync`

---

## Prerequisites

1. **Backend running**: `cd backend && make run` (port 8080)
2. **Frontend running**: `cd frontend && npm run dev` (Vite dev server, proxied to backend)
3. **Git branch**: `git checkout 005-vm-dashboard-ui-sync`

## Development Order

### Phase 1: Design System Foundation
1. Extend `frontend/src/index.css` with glassmorphism tokens and new component classes
2. Add new CSS variables for glass effects, glow states, and metric gauges

### Phase 2: Layout Restructuring
1. Restyle `Sidebar.tsx` — enhanced brand area, nav hover states, system footer
2. Restyle `Layout.tsx` TopBar — page title, contextual search, profile dropdown
3. Add responsive sidebar behavior (collapse on tablet, hide on mobile)

### Phase 3: Component Library
1. Create `VMCard.tsx` — extract and enhance VM card from Inventory
2. Create `StatusBadge.tsx` — reusable online/offline badge
3. Create `MetricGauge.tsx` — compact CPU/RAM gauge for cards
4. Restyle `ProvisioningBadge.tsx` — dark theme tokens
5. Extract `DeleteConfirmDialog.tsx` — standalone from Inventory
6. Restyle `CompactPowerActions.tsx` — match new card design

### Phase 4: Inventory Page Redesign
1. Rebuild `Inventory.tsx` — use new VMCard grid, staggered stats fetching
2. Wire per-card telemetry (online VMs only, staggered requests)
3. Apply React.memo to VMCard with custom comparator
4. Restyle empty state, search behavior, and modal triggers

### Phase 5: Dashboard & Secondary Pages
1. Restyle `DashboardView.tsx` — glassmorphism hero cards
2. Restyle `VMDetailView.tsx` — new design tokens
3. Restyle `AuditLogPage.tsx`, `KeysManagement.tsx`, `SettingsPage.tsx`
4. Restyle `Login.tsx` — dark theme

### Phase 6: Integration Testing & Validation
1. Verify all API endpoints work with redesigned components
2. Test all VM lifecycle actions (add, edit, terminal, power, delete)
3. Test responsive behavior at desktop/tablet/mobile breakpoints
4. Validate against reference image

## Key Files

| File | Purpose |
|------|---------|
| `frontend/src/index.css` | Design system — CSS custom properties and component classes |
| `frontend/src/components/Layout.tsx` | App shell — sidebar + topbar + content |
| `frontend/src/components/Sidebar.tsx` | Navigation sidebar |
| `frontend/src/pages/Inventory.tsx` | VM inventory grid (primary page) |
| `frontend/src/pages/DashboardView.tsx` | Dashboard overview with hero cards |
| `frontend/src/pages/VMDetailView.tsx` | VM detail with telemetry |
| `frontend/src/services/api.ts` | Axios instance with auth interceptor |
| `frontend/src/context/AuthContext.tsx` | Authentication state |

## Useful Commands

```bash
# Start backend
cd backend && make run

# Start frontend
cd frontend && npm run dev

# TypeScript check
cd frontend && npx tsc --noEmit

# Lint
cd frontend && npm run lint
```
