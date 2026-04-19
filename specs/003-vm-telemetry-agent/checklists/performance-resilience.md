# Unit Tests for Requirements: Performance & Resilience

**Purpose**: Validate the rigor of edge cases surrounding the disk-backed WAL, batching mechanisms, and network fault tolerance.
**Created**: 2026-04-19
**Feature**: specs/003-vm-telemetry-agent/spec.md

## Requirement Completeness

- [ ] CHK001 - Are the sync semantics for the disk-backed WAL (e.g., `fsync` per batch vs deferred sync) explicitly specified? [Completeness, Spec §FR-009]
- [ ] CHK002 - Is the backoff multiplier/strategy defined for retry attempts when the gRPC server is unreachable? [Completeness, Spec §Edge Cases]
- [ ] CHK003 - Are the CPU and memory consumption limits defined for the Go agent itself to prevent it from impacting host workloads? [Gap]

## Edge Case Coverage: WAL Exhaustion & Corruption

- [ ] CHK004 - Is the precise eviction policy defined when the WAL reaches 50MB (e.g., drop oldest batch, drop individual oldest events)? [Clarity, Spec §FR-009]
- [ ] CHK005 - Are requirements specified for how the agent detects and recovers from a corrupted WAL file following a kernel panic? [Coverage, Recovery Flow]
- [ ] CHK006 - Is the agent's behavior defined if the host filesystem becomes entirely read-only, preventing WAL writes? [Coverage, Exception Flow]
- [ ] CHK007 - Does the spec define behavior if the WAL directory lacks sufficient disk space *before* reaching the 50MB logical limit? [Coverage, Boundary Condition]

## Scenario Coverage: Transport & Throughput

- [ ] CHK008 - Are requirements defined for handling partial batch ingestions (e.g., 50 events ingested, connection drops mid-stream)? [Coverage, Exception Flow]
- [ ] CHK009 - Is the backend cleanup logic specified for identifying and terminating "zombie" gRPC streams during network flapping? [Coverage, Edge Case]
- [ ] CHK010 - Is there a defined timeout threshold for the agent to consider a batched payload transmission "failed"? [Completeness, Spec §FR-004]

## Acceptance Criteria Quality

- [ ] CHK011 - Can the 80% network overhead reduction [SC-004] be measured continuously, or is it a one-time benchmark requirement? [Measurability]
- [ ] CHK012 - Is the baseline for the 50% CPU drop [SC-002] explicitly documented so validation can be objectively verified? [Clarity]