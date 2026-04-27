# Implementation Plan: VM Dashboard UI Redesign & Backend Sync

**Branch**: `005-vm-dashboard-ui-sync` | **Date**: 2026-04-27 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `specs/005-vm-dashboard-ui-sync/spec.md`

## Summary

Redesign the VM management dashboard frontend to achieve a premium, dark-themed admin aesthetic matching the provided reference image. The work consists of two parallel tracks: (1) visual redesign of all UI components with glassmorphism panels, enhanced card layouts, and refined design tokens; and (2) frontend alignment with existing backend API contracts to ensure all available data (telemetry, provisioning status, audit events) is correctly consumed and displayed. No backend changes are required — the frontend adapts to existing endpoints.

## Technical Context

**Language/Version**: TypeScript 5.5 (strict mode), React 18.3, ES2020 target  
**Primary Dependencies**: Vite 8.0 (bundler), React Router DOM 6.30, Axios 1.15, Lucide React 1.8, Recharts 3.8, xterm 5.3  
**Storage**: N/A (frontend SPA; backend uses SQLite)  
**Testing**: Manual testing + TypeScript type-checking (`npx tsc --noEmit`)  
**Target Platform**: Web browser (Chrome/Firefox/Safari), responsive design  
**Project Type**: Single-page web application (Vite + React + TypeScript + Tailwind CSS)  
**Performance Goals**: < 3s time-to-interactive, < 100ms UI interactions, smooth rendering with 100 VMs  
**Constraints**: No backend API changes, existing Tailwind + CSS custom properties system  
**Scale/Scope**: ~14 components to restyle/create, ~7 pages, 1 CSS design system file

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| **1. Security first** | ✅ PASS | No credential handling changes. Existing AES encryption, Bearer auth, and audit logging preserved. No secrets exposed in UI state. |
| **2. Reliable backend integration** | ✅ PASS | Frontend components map directly to existing backend fields/actions. No unsupported behavior invented. |
| **3. Modern dashboard UX** | ✅ PASS | Redesign focuses on operational clarity and professional admin-dashboard experience using glassmorphism. |
| **4. Theme and visual consistency** | ✅ PASS | Dark-mode first design. Light mode is explicitly deferred to future work, as permitted by this principle. |
| **5. Real-time and telemetry strategy** | ✅ PASS | SSH terminal uses WebSocket correctly. HTTP polling for telemetry is documented in the spec as an accepted deviation since the backend lacks WebSocket metrics endpoints. |
| **6. Code quality and maintainability** | ✅ PASS | TypeScript strict mode enabled. Component-driven architecture with explicit interfaces. All error paths surface user-actionable messages. |
| **7. Accessibility and responsiveness** | ✅ PASS | UI remains responsive across mobile, tablet, and desktop breakpoints. Hover and focus states are clearly defined. |
| **8. Performance awareness** | ✅ PASS | React.memo used for 100+ VM lists to prevent unnecessary re-renders. Staggered requests implemented for telemetry to avoid API flooding. |

**Post-Phase 1 Re-check**: All gates still pass. No new violations introduced by design artifacts.

## Project Structure

### Documentation (this feature)

```text
specs/005-vm-dashboard-ui-sync/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
backend/
├── cmd/
├── internal/
│   ├── api/
│   ├── models/
│   └── monitor/

frontend/
├── src/
│   ├── components/
│   ├── pages/
│   ├── services/
│   ├── types/
│   └── index.css
```

**Structure Decision**: Web application layout. Frontend work will be confined to `frontend/src/` with no changes to `backend/`.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

*(No violations - all principles passed successfully)*
