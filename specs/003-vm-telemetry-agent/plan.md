# Implementation Plan: VM Telemetry Agent

**Branch**: `002-v2-enterprise-upgrade` | **Date**: 2026-04-19 | **Spec**: [specs/003-vm-telemetry-agent/spec.md](spec.md)
**Input**: Feature specification from `/specs/003-vm-telemetry-agent/spec.md`

## Summary

Migrate from SSH-based agentless polling to a dedicated Go-based VM agent with a zero-trust identity lifecycle (CSR/mTLS). The agent will stream hardware metrics and structured logs to the backend in batched payloads, backed by a 50MB disk-backed Write-Ahead Log (WAL) to ensure data integrity during network disconnects.

## Technical Context

**Language/Version**: Go 1.20+
**Primary Dependencies**: gRPC, Protocol Buffers, standard library crypto/tls
**Storage**: Disk-backed WAL (Agent, 50MB limit), SQLite (Backend)
**Testing**: Go `testing` package (unit/integration)
**Target Platform**: Standard Linux (Ubuntu/Debian, RHEL/CentOS), Containerized Workloads
**Project Type**: Backend Service & Daemon Agent
**Performance Goals**: 80% reduction in network overhead via 500ms / 100-event batched streams
**Constraints**: Zero-Trust provisioning (private keys never cross the network), WAL strictly capped at 50MB
**Scale/Scope**: Enterprise VM fleets requiring real-time observability

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Security First**: PASS. Zero-trust CSR workflow explicitly prevents private keys from traversing the network. mTLS enforces strict mutual authentication.
- **Real-time Efficiency**: PASS. Batched gRPC streaming optimizes network overhead, adhering to the spirit of efficient, low-latency communication over continuous polling.
- **Code Quality**: PASS. Go's strict typing and interface-driven design will manage the gRPC contracts and WAL components.
- **Modern UI**: N/A. (Backend and agent infrastructure only).

## Project Structure

### Documentation (this feature)

```text
specs/003-vm-telemetry-agent/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
│   └── telemetry.proto
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
backend/
├── cmd/
│   └── agent/
│       └── main.go       # New Agent Entrypoint
├── internal/
│   ├── api/
│   │   └── grpc/         # Protobuf generated code and gRPC server implementation
│   ├── agent/            # Agent specific logic
│   │   ├── bootstrap/    # CSR and Identity lifecycle
│   │   ├── telemetry/    # Hardware sampling and log aggregation
│   │   └── wal/          # Disk-backed Write-Ahead Log implementation
│   ├── monitor/          # (Existing) To be updated/deprecated
│   ├── ssh/              # (Existing) To be updated/deprecated
│   └── ca/               # Internal Certificate Authority for signing CSRs
├── scripts/
│   └── provisioning/
│       └── deploy_agent.sh # Refactored deployment script
```

**Structure Decision**: The agent codebase will live within the existing `backend/` module as a new binary (`cmd/agent`), sharing protobuf definitions and core utility libraries with the main server, but compiling independently for deployment to target VMs.

## Execution Phases

### Phase 1: Protocol Contracts
Define the gRPC Protobufs (`.proto`) supporting the Certificate Signing Request (CSR) lifecycle and client-side batched streaming for both metrics and structured logs.
- Create `telemetry.proto`.
- Generate Go code using `protoc`.

### Phase 2: Backend Security & Ingestion
Configure the backend gRPC server.
- Implement the Certificate Authority (CA) logic to automatically validate and sign agent CSRs.
- Establish the ingestion handlers for batched telemetry, routing the streamed batches into the existing database models.

### Phase 3: Agent Bootstrap & Identity
Implement the Go agent's initialization sequence.
- Add local private key generation.
- Execute the CSR flow against the backend gRPC endpoint.
- Ensure secure local storage of the resulting mTLS certificate and private key.

### Phase 4: Agent Telemetry & WAL Resilience
Implement telemetry collection and resilience.
- Build hardware metric sampling (CPU, RAM, Disk, Network) and structured log aggregation.
- Construct the 50MB disk-backed Write-Ahead Log (WAL) to buffer batched payloads during network disconnection events.

### Phase 5: Automated Provisioning Pipeline
Refactor the SSH-based deployment scripts.
- Modify `scripts/provisioning/deploy_agent.sh` (or create it) to exclusively transfer the compiled Go binary and a systemd unit file to the target VMs.
- Remove any logic that generated or transferred client keys over SSH.

### Phase 6: Deprecation
Outline the specific files and functions required to safely strip the legacy SSH polling mechanisms from the backend codebase.
- Safely remove polling loops from `backend/internal/ssh` and `backend/internal/monitor`.
- Retain SSH exclusively for direct terminal access and explicit administrative commands as mandated.