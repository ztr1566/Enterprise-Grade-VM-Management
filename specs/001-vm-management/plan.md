# Implementation Plan: Web-Based VM Management Platform

**Branch**: `001-vm-management` | **Date**: 2026-04-17 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-vm-management/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Web-Based VM Management Platform - Golang Backend with React Frontend, utilizing agentless SSH over WebSockets, secure credential management, and explicit credential audit logging.

## Technical Context

**Language/Version**: Go (Golang) v1.21+, TypeScript (Strict)
**Primary Dependencies**: 
- **Frontend**: React, Vite, Tailwind CSS, xterm.js
- **Backend**: gorilla/websocket, golang.org/x/crypto/ssh, mattn/go-sqlite3, golang-jwt/jwt, uber-go/zap (for Audit Logging)
**Storage**: SQLite
**Testing**: Go standard `testing` package (Backend) / Vitest + React Testing Library (Frontend).
**Target Platform**: Linux server (Host), modern web browsers (Client).
**Project Type**: web-service
**Performance Goals**: Support 50 concurrent VM SSH/WebSocket sessions per instance with <150ms keystroke latency, and metrics refreshing every 5s.
**Constraints**: Agentless SSH only. All credential access MUST be explicitly audit logged. WebSocket exponential backoff must be implemented on the frontend.
**Scale/Scope**: Up to 50 concurrent target VMs simultaneously monitored and accessed via WebSockets.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **Security First**: AES-256-GCM for credential storage. Explicit audit logging layer added (`uber-go/zap`) for tracking all credential access.
- [x] **Real-time Efficiency**: All metrics and SSH sessions piped exclusively via WebSockets (`gorilla/websocket`). Frontend implements exponential backoff on disconnect.
- [x] **Code Quality**: Strict TypeScript for frontend. Go error handling centralizes logging and presents actionable messages to the UI.
- [x] **Modern UI**: React component-based UI with Tailwind CSS dark-mode first design.

## Project Structure

### Documentation (this feature)

```text
specs/001-vm-management/
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
│   └── server/           # Main Go application entrypoint
├── internal/
│   ├── models/           # SQLite entity models
│   ├── ssh/              # crypto/ssh connection pooling
│   ├── api/              # HTTP REST routes
│   ├── ws/               # WebSocket dispatchers
│   ├── auth/             # JWT authentication middleware
│   └── audit/            # uber-go/zap structured audit logging
└── go.mod

frontend/
├── src/
│   ├── components/
│   ├── pages/
│   ├── services/
│   └── utils/            # Exponential backoff WS helpers
└── tests/
```

**Structure Decision**: Option 2: Web application. The project is split down the middle with distinct Frontend (React Vite project) and Backend (Golang API/WS router) domains to cleanly isolate the high-performance SSH pooling logic and structured audit logging from the UI rendering cycles.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| None | N/A | N/A |
