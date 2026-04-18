<!--
  Sync Impact Report
  ===================
  Version change: N/A (initial) → 1.0.0
  Modified principles: N/A (initial ratification)
  Added sections:
    - Core Principles (4 principles: Security First, Real-time Efficiency,
      Code Quality, Modern UI)
    - Technology Constraints
    - Development Workflow
    - Governance
  Removed sections:
    - Principle 5 placeholder (user specified exactly 4 principles)
  Templates requiring updates:
    - .specify/templates/plan-template.md ✅ no changes needed
    - .specify/templates/spec-template.md ✅ no changes needed
    - .specify/templates/tasks-template.md ✅ no changes needed
  Follow-up TODOs: None
-->

# Infrastructure Dashboard Constitution

## Core Principles

### I. Security First

- All SSH credentials MUST be stored using encrypted local storage.
  Plain-text passwords MUST NOT be persisted at rest or in transit
  within the application.
- The system MUST implement a secure key management layer that
  handles SSH key passphrases, host key verification, and credential
  lifecycle (creation, rotation, revocation).
- Secrets MUST NOT appear in application logs, error messages, or
  client-side state accessible via browser developer tools.
- Every credential access MUST be auditable — the system MUST record
  which user accessed which credential and when.

**Rationale**: Infrastructure management tools are high-value targets.
A single credential leak can compromise entire server fleets. Secure
defaults are non-negotiable.

### II. Real-time Efficiency

- Live dashboard metrics (CPU, memory, disk, network) MUST be
  delivered exclusively via WebSocket connections. HTTP polling
  MUST NOT be used for any real-time data stream.
- SSH terminal sessions MUST use WebSocket-backed channels to
  provide bidirectional, low-latency communication between the
  browser and remote hosts.
- WebSocket connections MUST implement automatic reconnection with
  exponential backoff and provide clear connection-state indicators
  to the user.
- The system MUST gracefully degrade when a WebSocket connection
  drops — buffered data MUST NOT be silently discarded.

**Rationale**: Polling introduces unnecessary latency and server load.
WebSockets provide the persistent, full-duplex channel required for
interactive terminals and sub-second metric refresh rates.

### III. Code Quality

- All source code MUST use strict typing (TypeScript `strict` mode
  or equivalent). Implicit `any` types MUST NOT be permitted.
- The codebase MUST follow a modular architecture where each
  concern (SSH management, credential storage, metrics collection,
  UI rendering) is encapsulated in its own module with explicit
  public interfaces.
- Comprehensive error handling is MANDATORY for:
  - Failed SSH connection attempts (authentication failures,
    unreachable hosts, refused connections).
  - Connection timeouts (configurable per-host thresholds).
  - Network interruptions (mid-session disconnects, WebSocket
    failures).
- Every error path MUST surface a user-actionable message — raw
  stack traces or cryptic error codes MUST NOT reach the UI.

**Rationale**: Infrastructure tools are operated under pressure
(outage scenarios). Strict types prevent runtime surprises, modular
code enables safe iteration, and clear error handling reduces
mean-time-to-resolution.

### IV. Modern UI

- The UI MUST be built using a component-driven framework (e.g.,
  React, Vue, Svelte) to ensure reusable, testable interface
  elements.
- A dark-mode-first design MUST be the default theme, with an
  optional light-mode toggle. Dark mode is the primary design
  target and MUST NOT be an afterthought.
- The interface MUST feel responsive and professional, suitable
  for extended use during infrastructure management sessions —
  clean typography, consistent spacing, and subdued color palettes
  that reduce eye strain.
- All interactive elements (buttons, terminal panes, metric graphs)
  MUST provide immediate visual feedback on user actions (hover
  states, loading indicators, transition animations).

**Rationale**: DevOps and SRE engineers spend hours in dashboards
during incidents. A polished, dark-mode-first UI with clear visual
hierarchy directly impacts operator efficiency and reduces fatigue.

## Technology Constraints

- **Transport layer**: WebSocket (via `ws` or `socket.io`) for all
  real-time features. REST/HTTP endpoints are permitted only for
  CRUD operations, authentication flows, and initial page loads.
- **Credential storage**: An encrypted local store (e.g., `keytar`,
  OS keychain integration, or AES-256-GCM encrypted file) MUST be
  used. Database-backed credential stores MUST encrypt at the field
  level, not merely at-rest disk encryption.
- **Type system**: TypeScript with `strict: true` in `tsconfig.json`
  is REQUIRED for all JavaScript/TypeScript codebases. If another
  language is chosen, an equivalent strict-typing configuration
  MUST be enabled.
- **SSH library**: A well-maintained library (e.g., `ssh2` for
  Node.js) MUST be used. Direct OS-level `ssh` command spawning
  MUST NOT be the primary connection mechanism.

## Development Workflow

- **Branch strategy**: All feature work MUST occur on dedicated
  feature branches. Direct commits to `main` are prohibited.
- **Code review**: Every pull request MUST pass automated linting,
  type-checking, and at least one peer review before merge.
- **Error handling audit**: Each PR that introduces new SSH or
  WebSocket paths MUST include explicit error-handling coverage
  for timeout, disconnect, and authentication-failure scenarios.
- **Security review**: Any change to credential storage, key
  management, or authentication logic MUST receive a dedicated
  security-focused review.

## Governance

- This constitution is the supreme authority for project standards.
  When a conflict arises between this document and any other
  practice, specification, or implementation plan, this document
  prevails.
- **Amendment procedure**: Amendments MUST be proposed via a pull
  request modifying this file. The PR MUST include a rationale,
  the version bump classification (MAJOR/MINOR/PATCH), and an
  impact assessment on existing code.
- **Versioning policy**: The constitution follows Semantic
  Versioning — MAJOR for principle removals or redefinitions,
  MINOR for new principles or material expansions, PATCH for
  clarifications and wording fixes.
- **Compliance review**: All specifications (`spec.md`),
  implementation plans (`plan.md`), and task lists (`tasks.md`)
  MUST be validated against this constitution at creation time.
  The plan-template Constitution Check section enforces this gate.

**Version**: 1.0.0 | **Ratified**: 2026-04-17 | **Last Amended**: 2026-04-17
