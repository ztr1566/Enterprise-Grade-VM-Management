# Frontend UI Contract: VM Dashboard

**Branch**: `005-vm-dashboard-ui-sync`  
**Date**: 2026-04-27  

---

## Component Hierarchy Contract

This document defines the component structure, props contracts, and rendering responsibilities for the redesigned frontend.

### Layout Components

#### `Layout.tsx`
- **Renders**: Sidebar + TopBar + content area (Outlet)
- **Props**: None (root layout, uses React Router Outlet)
- **State**: `searchTerm: string`, `serverCount: number`
- **Passes to children via Outlet context**: `{ searchTerm: string }`
- **API dependency**: `GET /api/dashboard/stats` (for server count badge)

#### `Sidebar.tsx`
- **Renders**: Brand logo, navigation links, system status footer
- **Props**: None (uses `useLocation` for active state)
- **Navigation items**: Dashboard, VMs, Audit Log, Key Vault, Settings
- **Responsive**: Collapsed on tablet, hidden on mobile with hamburger toggle

#### `TopBar` (within Layout)
- **Renders**: Page title, contextual search, user profile dropdown
- **State**: `searchTerm`, `dropdownOpen`
- **Contextual behavior**: Search input visible only on `/vms` route

### Inventory Components

#### `Inventory.tsx` (page)
- **Renders**: Page header + VM card grid + modals
- **State**: `vms: VM[]`, `vmStats: Map<string, VMStats>`, modal states
- **API**: `GET /api/vms` (list), `GET /api/vms/{id}/stats` (per online VM, staggered)
- **Polling**: 5s when pending VMs, 10s otherwise
- **Search**: Filters `vms` by name, host, or tags using `searchTerm` from context

#### `VMCard.tsx` (new component — extracted from Inventory)
- **Props**:
  ```typescript
  interface VMCardProps {
    vm: VM;
    stats?: VMStats;
    onTerminal: (vm: VM) => void;
    onEdit: (vm: VM) => void;
    onDelete: (vm: VM) => void;
    onPowerAction: (vmId: string, action: string) => void;
    onRetryProvisioning: (vmId: string) => void;
  }
  ```
- **Renders**: Server icon, VM name, host, status badge, provisioning badge, CPU/RAM gauges (if stats), action buttons
- **Memoized**: `React.memo` with custom comparator on `vm.id`, `vm.status`, `vm.provisioning_status`, `stats?.cpu`, `stats?.ram`

#### `StatusBadge.tsx` (new component)
- **Props**: `{ status: "online" | "offline" }`
- **Renders**: Pill badge with colored dot + text using `status-badge` CSS classes

#### `ProvisioningBadge.tsx` (restyle existing)
- **Props**: `{ status: string; onRetry?: () => void }`
- **Change**: Use `prov-badge` CSS classes instead of light-mode Tailwind utilities

#### `MetricGauge.tsx` (new component)
- **Props**: `{ label: string; value: number; color: "blue" | "violet" | "amber" }`
- **Renders**: Compact progress bar with percentage label for CPU/RAM on cards
- **Behavior**: Shows "—" when value is undefined/null (no telemetry data)

#### `CompactPowerActions.tsx` (restyle existing)
- **Props**: unchanged
- **Change**: Visual restyle to match new card design

### Shared Components

#### `DeleteConfirmDialog.tsx` (extract from Inventory)
- **Props**: `{ vm: VM; onConfirm: () => void; onCancel: () => void }`
- **Change**: Extract as standalone component for reusability

#### `AddVmModal.tsx` (restyle existing)
- **Props**: unchanged
- **Change**: Visual restyle to match new design system

### Page Components (restyle)

#### `DashboardView.tsx` — Restyle hero cards with glassmorphism
#### `VMDetailView.tsx` — Restyle with new design tokens
#### `AuditLogPage.tsx` — Restyle table and pagination
#### `KeysManagement.tsx` — Restyle cards and forms
#### `Login.tsx` — Restyle to match dark theme
#### `SettingsPage.tsx` — Restyle settings cards

---

## CSS Design Token Contract

### New tokens to add to `index.css`:

```css
/* Glassmorphism */
--glass-bg: rgba(17, 24, 39, 0.6);
--glass-border: rgba(255, 255, 255, 0.08);
--glass-blur: 12px;

/* Card glow effects */
--glow-blue: 0 0 20px rgba(59, 130, 246, 0.15);
--glow-emerald: 0 0 20px rgba(16, 185, 129, 0.15);
--glow-red: 0 0 20px rgba(239, 68, 68, 0.15);
```

### New CSS classes:

```css
.glass-panel { ... }         /* Glassmorphism container */
.glass-card { ... }          /* Glassmorphism card variant */
.metric-gauge { ... }        /* Compact inline metric bar */
.metric-gauge-fill { ... }   /* Metric bar fill */
.vm-card { ... }             /* New VM card style */
.vm-card-actions { ... }     /* Card action button group */
```
