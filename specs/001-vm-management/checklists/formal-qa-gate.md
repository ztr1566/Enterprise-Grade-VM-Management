# Specification Quality Checklist: Formal QA & Security Gate

**Purpose**: Rigorous validation of requirements completeness, edge-case coverage, and deployment prerequisites before implementation.
**Created**: 2026-04-17
**Feature**: Web-Based VM Management Platform

## Security & RBAC Requirements

- [ ] CHK001 - Are the boundary conditions for RBAC "Viewer" vs "Admin" roles explicitly defined for all API endpoints? [Completeness, Gap]
- [ ] CHK002 - Is the maximum operational lifespan or rotation strategy for the AES-256-GCM master key specified? [Clarity, Edge Case]
- [ ] CHK003 - Are data protection requirements documented for handling credentials in-memory prior to zeroing them out during SSH execution? [Coverage, Security]
- [ ] CHK004 - Are authentication rate-limiting or brute-force mitigation thresholds quantified? [Measurability, Spec §API]

## Architecture & Real-Time Constraints

- [ ] CHK005 - Are failure state requirements defined for when the WebSocket connection to the client drops mid-terminal session? [Exception Flow, Gap]
- [ ] CHK006 - Do requirements specify the exact rollback or buffering behavior when agentless SSH telemetry parsing fails (e.g., target uses an unsupported `top` version)? [Edge Case, Coverage]
- [ ] CHK007 - Are the metric reporting latency thresholds consistently defined between the telemetry contract (5s) and internal processing rules? [Consistency]
- [ ] CHK008 - Are requirements clear on how concurrent administrator sessions viewing the *same* VM terminal are handled? [Coverage, Concurrency]

## UX, Terminal & Interface

- [ ] CHK009 - Is the "< 150ms Terminal Latency" requirement objectively measurable strictly over the network, or inclusive of rendering? [Measurability]
- [ ] CHK010 - Are visual degradation requirements specified for when the terminal connection degrades above 150ms? [Edge Case, Spec §NFR]
- [ ] CHK011 - Does the spec define the exact visual fallbacks or error messages if the requested service (e.g., `nginx.service`) doesn't exist on the target? [Completeness]

## Deployment & Infrastructure (SQLite/Network)

- [ ] CHK012 - Are deployment limits for SQLite concurrent writes specified during peak telemetry pushes (50 VMs at 5s)? [Completeness, Scaling]
- [ ] CHK013 - Are network firewall prerequisites (outbound rules, WebSocket inbound rules) explicitly documented for the deployment environment? [Coverage, Deployment]
- [ ] CHK014 - Do requirements define disaster recovery or backup procedures for the SQLite instance holding the encrypted master credentials? [Gap, Recovery Flow]
- [ ] CHK015 - Are resource allocation requirements (vCPUs, RAM limits) specified for the Node.js server hosting up to 50 active PTY streams? [Clarity, Constraints]

## Functional Acceptance Criteria

- [ ] CHK016 - Can the 60-second VM onboarding criteria be tested independently of network latency variations? [Measurability, Spec §SC-001]
- [ ] CHK017 - Are validation rules documented for the `host` string formats (IPv4 vs IPv6 vs Hostname)? [Clarity, Data Model]
- [ ] CHK018 - Do the quick-action timeout requirements (3s limit) specify if the timeout cancels the SSH transmission or just the UI loading spinner? [Clarity]
