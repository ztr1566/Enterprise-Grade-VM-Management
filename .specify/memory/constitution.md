<!--
  Sync Impact Report
  ===================
  Version change: 1.0.0 → 2.0.0
  Modified principles: Complete rewrite. Replaced 4 technical principles with 8 product/UX-aligned principles.
  Added sections:
    - Explicit product direction
    - Decision rules
    - Scope boundaries
    - Review policy
  Removed sections:
    - Technology Constraints
    - Development Workflow
  Templates requiring updates:
    - .specify/templates/plan-template.md ⚠ pending manual review (Constitution Check section is generic, so it doesn't strictly break, but the new principles should be used)
    - .specify/templates/spec-template.md ✅ no changes needed
    - .specify/templates/tasks-template.md ✅ no changes needed
  Follow-up TODOs: None
-->

# Project Constitution

This constitution defines the stable architectural, product, UX, and engineering principles for the Enterprise-Grade-VM-Management project.

## Purpose
The project is a modern VM management dashboard focused on clarity, operational efficiency, and reliable frontend/backend alignment. All future specs, plans, and tasks must follow this constitution unless a deliberate exception is documented.

## Core principles

### 1. Security first
- Preserve authentication, authorization, and session handling as strict requirements.
- Do not weaken existing security controls for convenience.
- Any new feature that touches sensitive operations must preserve least privilege and explicit user intent.

### 2. Reliable backend integration
- The frontend must stay aligned with the backend contract.
- UI components must reflect actual backend fields, states, and actions.
- If backend capabilities change, the frontend must be updated to match without inventing unsupported behavior.

### 3. Modern dashboard UX
- The product should present a polished, professional admin-dashboard experience.
- Use clear hierarchy, responsive layout, readable typography, and predictable interactions.
- The UI should favor operational clarity over decorative complexity.

### 4. Theme and visual consistency
- The default experience may be dark-mode-first when it better fits the product direction.
- Light mode is optional and may be introduced in a future feature if required.
- Visual style must remain consistent across pages, components, and states.

### 5. Real-time and telemetry strategy
- Live metrics should use the most appropriate transport supported by the backend.
- WebSocket is preferred for true real-time streams when available.
- If the backend does not provide WebSocket telemetry, controlled HTTP polling is acceptable as an interim or constrained solution.
- Any deviation from WebSocket-first real-time delivery must be documented in the relevant spec.

### 6. Code quality and maintainability
- Prefer reusable components, clear boundaries, and type-safe implementation.
- Avoid duplication between UI layers, data-mapping layers, and API helpers.
- Keep code easy to extend, test, and review.

### 7. Accessibility and responsiveness
- The UI must remain accessible and responsive across common desktop and mobile breakpoints.
- Interactive elements must be keyboard reachable and have clear focus states.
- Text, contrast, spacing, and controls must remain usable in both light and dark themes where supported.

### 8. Performance awareness
- The dashboard should remain responsive with realistic inventory sizes.
- Optimize rendering, state updates, and data refresh behavior where needed.
- Avoid unnecessary re-renders and expensive UI work in common flows.

## Explicit product direction
- The VM dashboard should be redesigned to match approved visual references when a new UI refresh is requested.
- Frontend updates must stay synchronized with backend enhancements already completed.
- UI changes must not break existing workflows unless the change is explicitly intended and documented.

## Decision rules
- If a feature conflicts with this constitution, the constitution takes priority.
- If the constitution no longer reflects the real product direction, update the constitution first before writing a new feature spec.
- If a rule is intentionally bypassed for a specific feature, the exception must be documented in that feature’s spec under constraints or accepted deviations.

## Scope boundaries
- This constitution defines project-wide principles, not implementation details.
- Specific UI layouts, endpoints, component names, and task breakdowns belong in feature specs and plans.
- Temporary or feature-specific exceptions must not silently become permanent project rules.

## Review policy
- Revisit this constitution whenever the product direction, architecture, or platform assumptions change.
- Keep it short, stable, and enforceable.
- Remove outdated rules instead of carrying contradictory guidance forward.

**Version**: 2.0.0 | **Ratified**: 2026-04-17 | **Last Amended**: 2026-04-27
