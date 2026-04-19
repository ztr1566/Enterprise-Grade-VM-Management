# Unit Tests for Requirements: Protocol & Contracts

**Purpose**: Validate the completeness, clarity, and consistency of the Protobuf contracts and gRPC schema definitions before implementation.
**Created**: 2026-04-19
**Feature**: specs/003-vm-telemetry-agent/spec.md

## Requirement Completeness

- [ ] CHK001 - Are the precise data types (e.g., int64, double) explicitly specified for all fields within the MetricPayload contract? [Completeness]
- [ ] CHK002 - Is the maximum payload size per gRPC streaming batch defined to prevent backend memory exhaustion? [Gap, Spec §FR-004]
- [ ] CHK003 - Are the valid enum values or string formats for the LogPayload `severity` field explicitly documented? [Completeness, Spec §FR-005]
- [ ] CHK004 - Does the contract define the required metadata fields for the CSRRequest (e.g., SANs, CN structure) beyond just `vm_id`? [Gap]

## Requirement Clarity

- [ ] CHK005 - Is the unit of measurement explicitly defined for the CPU and Disk metrics (e.g., percentage vs raw ticks)? [Clarity, Spec §FR-003]
- [ ] CHK006 - Is the format of the `timestamp` field unambiguously defined (e.g., Unix epoch seconds vs milliseconds)? [Clarity, Spec §FR-003, §FR-005]
- [ ] CHK007 - Is the backend's acknowledgment behavior clearly specified (e.g., ack per batch vs ack per stream closure)? [Clarity, Spec §FR-004]

## Scenario & Edge Case Coverage

- [ ] CHK008 - Are error response codes defined for when the backend rejects a malformed telemetry batch? [Coverage, Exception Flow]
- [ ] CHK009 - Is the expected behavior defined for schema version mismatches between older agents and a newer backend? [Coverage, Edge Case]
- [ ] CHK010 - Are there defined limits on the length of the `message` string within a LogPayload to prevent single-event payload bloat? [Coverage, Boundary Condition]

## Consistency & Dependencies

- [ ] CHK011 - Does the `vm_id` field in the gRPC payloads consistently map to the existing primary key format used by the backend database? [Consistency]
- [ ] CHK012 - Are the Protobuf definitions consistent with the established data models required by `backend/internal/monitor`? [Dependency]