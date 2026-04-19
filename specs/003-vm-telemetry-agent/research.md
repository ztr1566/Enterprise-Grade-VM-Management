# Phase 0: Research & Decisions

## Telemetry Transmission (Performance)
- **Decision**: Use batched client-side streaming via gRPC.
- **Rationale**: Replaces continuous streaming of individual log lines. Batching at 500ms intervals or 100-event chunks significantly mitigates network overhead (target 80% reduction).
- **Alternatives**: HTTP polling (rejected by constitution), continuous individual gRPC streams (rejected due to overhead).

## Local Buffering Resilience (Data Integrity)
- **Decision**: Implement a 50MB disk-backed Write-Ahead Log (WAL) on the agent.
- **Rationale**: Ensures data integrity during sudden power loss or kernel panics. Allows buffering of batched payloads during network disconnection.
- **Alternatives**: In-memory buffering (rejected due to data loss risk on panic/restart).

## Security & Identity Lifecycle (Zero-Trust)
- **Decision**: Agent generates its own private key and executes a Certificate Signing Request (CSR) flow.
- **Rationale**: Eliminates the risk of distributing private keys over SSH. The backend acts as a CA to automatically validate and sign agent CSRs.
- **Alternatives**: Backend generates certs and SCPs them (rejected due to zero-trust mandate).