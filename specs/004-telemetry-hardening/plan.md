# Implementation Plan: Telemetry Agent Hardening

**Branch**: `004-telemetry-hardening` | **Date**: 2026-04-20 | **Spec**: [specs/004-telemetry-hardening/spec.md](spec.md)
**Input**: Feature specification from `/specs/004-telemetry-hardening/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Implement security, resilience, and protocol hardening for the VM Telemetry Agent. This includes 30-day mTLS certificate lifecycle with auto-renewal, OTT-based CSR authorization, a robust disk-backed WAL with fsync and corruption recovery, strict resource governance (100MB RAM/10% CPU), and 4MB gRPC payload enforcement.

## Technical Context

**Language/Version**: Go 1.25.0
**Primary Dependencies**: gRPC, Protocol Buffers, `crypto/ecdsa`, `crypto/tls`, `github.com/shirou/gopsutil/v3`
**Storage**: Disk-backed WAL (Agent, 50MB limit), SQLite (Backend)
**Testing**: Go `testing` package (unit/integration)
**Target Platform**: Standard Linux (Ubuntu/Debian, RHEL/CentOS), Containerized Workloads
**Project Type**: Backend Service & Daemon Agent
**Performance Goals**: 100MB RAM max, 10% CPU max, fsync every 10th batch
**Constraints**: Zero-Trust provisioning (keys never cross network), 4MB max payload, 30-day cert TTL, 10-minute OTT validity
**Scale/Scope**: Enterprise VM fleets (100+ concurrent agents)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Security First**: PASS. OTT for CSR, ECDSA P-256 keys with 0600 permissions, CRL for revocation, and 30-day mTLS certificate auto-renewal perfectly align with zero-trust principles.
- **Real-time Efficiency**: PASS. Exponential backoff with jitter prevents thundering herds. Backpressure via resource governance prevents the agent from starving host workloads.
- **Code Quality**: PASS. Go strict types, structured logging, and health endpoints fulfill observability and maintainability requirements.
- **Modern UI**: N/A. (Backend and agent infrastructure only).

## Project Structure

### Documentation (this feature)

```text
specs/004-telemetry-hardening/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

```text
backend/
├── cmd/
│   ├── agent/
│   └── server/
├── internal/
│   ├── agent/
│   │   ├── telemetry/
│   │   ├── wal/
│   │   └── governance/   # New: Resource limits
│   ├── api/
│   │   └── grpc/
│   │       ├── interceptors/ # New: Payload limits
│   │       └── telemetry/
│   ├── crypto/           # Updated: ECDSA/CSR logic
│   └── monitor/
└── tests/
    └── integration/
```

**Structure Decision**: The project uses the existing Go backend architecture. We will introduce a new `governance` package for the agent resource monitoring, and update `api/grpc/interceptors` for payload enforcement. The `wal` and `crypto` packages will be hardened.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| N/A | N/A | N/A |
