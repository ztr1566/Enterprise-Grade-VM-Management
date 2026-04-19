# Unit Tests for Requirements: Security & Zero-Trust

**Purpose**: Validate the strict enforcement and edge-case resilience of the mTLS, identity lifecycle, and zero-trust boundaries.
**Created**: 2026-04-19
**Feature**: specs/003-vm-telemetry-agent/spec.md

## Requirement Completeness

- [ ] CHK001 - Are the cryptographic algorithms and key lengths (e.g., ECDSA P-256, RSA-4096) specified for the agent's private key generation? [Completeness, Spec §FR-001]
- [ ] CHK002 - Are the file permission requirements (e.g., `0600`) documented for the locally generated private key? [Completeness, Spec §FR-001]
- [ ] CHK003 - Is the expiration lifecycle/TTL explicitly defined for the backend-issued mTLS certificates? [Gap]
- [ ] CHK004 - Are the authorization criteria for signing a CSR explicitly defined (e.g., one-time token, pre-shared key, IP allowlist)? [Gap, Spec §FR-001]

## Scenario & Edge Case Coverage

- [ ] CHK005 - Is the agent's recovery behavior defined if it reboots and its private key is missing or corrupted? [Coverage, Recovery Flow]
- [ ] CHK006 - Are requirements defined for how the agent handles an expiring mTLS certificate when the backend CA is unreachable? [Coverage, CA Outage]
- [ ] CHK007 - Is the backend's behavior specified when it receives a CSR from a `vm_id` that already possesses a valid, unexpired certificate? [Coverage, Edge Case]
- [ ] CHK008 - Are requirements defined for revoking compromised agent certificates (e.g., CRL or OCSP)? [Gap, Exception Flow]

## Requirement Clarity & Consistency

- [ ] CHK009 - Is the exact mechanism preventing the provisioning script from accessing the generated key clearly articulated? [Clarity, Spec §FR-001]
- [ ] CHK010 - Do the zero-trust requirements consistently align with the project's "Security First" constitutional principle regarding auditable credential access? [Consistency]
- [ ] CHK011 - Are the TLS version requirements (e.g., TLS 1.3 only) explicitly specified for the gRPC connection? [Clarity, Spec §FR-004]