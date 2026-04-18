# Feature Specification: V2 Enterprise Upgrade

**Feature Branch**: `003-v2-enterprise-upgrade`  
**Created**: 2026-04-18  
**Status**: Draft  
**Input**: User description: "V2 Enterprise Upgrade: Zero-touch sudoers provisioning for agentless security, restructure UI into a multi-page dashboard, and add advanced network/process telemetry."

## Clarifications

### Session 2026-04-18
- Q: How does the provisioning engine authenticate its first sudo call without violating the no-password constraint? → A: The SSH user provided during registration MUST already have either root access or passwordless sudo. If they don't, provisioning fails with a clear error instructing the admin to configure initial access.
- Q: How does the platform handle sudoers drift (rules manually removed after provisioning)? → A: No active drift detection. When a privileged command fails with a permission error, the error is surfaced to the user alongside a "Re-provision" button.
- Q: Which network interface(s) are monitored for telemetry on multi-interface VMs? → A: Aggregate traffic across all non-loopback, non-virtual interfaces. Exclude `lo`, `docker0`, `veth*`, and `br-*`.

## Architectural Constraints

The following constraints were explicitly defined by the project owner and MUST be adhered to during planning and implementation. They represent non-negotiable security and design decisions.

1. **Security First**: No sudo passwords shall ever be transmitted from the frontend to the backend. There must be no password input dialogs, no API endpoints accepting password payloads, and no credential caching for sudo operations. All privileged command execution must be pre-authorized on the target VM itself.
2. **Secure Provisioning Engine**: The backend must include a provisioning engine that automatically configures newly registered VMs for secure, passwordless execution of only the specific privileged commands the platform requires.
3. **Embedded Provisioning Script**: The backend must bundle a provisioning shell script at build time (located at `scripts/provisioning/setup_vm_sudoers.sh`) using Go's `//go:embed` directive, ensuring the script is version-controlled and tamper-resistant.
4. **Strict Sudoers Rules**: The provisioning script must inject NOPASSWD rules into the target VM's `/etc/sudoers.d/` directory, scoped exclusively to the `systemctl` and `journalctl` commands required by the platform. No blanket `NOPASSWD: ALL` rules are permitted.
5. **Safety Validation**: The provisioning script must validate the generated sudoers file syntax using `visudo -c` before applying it to the live system, preventing administrator lockouts from malformed rules.
6. **Automatic Execution**: The provisioning process must execute automatically over SSH immediately after a new server is registered in the platform, with no manual intervention required from the administrator.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Zero-Touch Secure Provisioning (Priority: P1) 🎯 MVP

As a system administrator, I want the platform to automatically configure newly registered servers with the minimum required sudo permissions so that all privileged platform operations (service management, log streaming, power actions) work immediately after registration without me having to manually edit sudoers files or store root passwords in the platform.

**Why this priority**: The V1 platform assumes all SSH users have passwordless sudo access, which either fails on hardened servers or forces administrators into insecure configurations. By automating secure provisioning at registration time, the platform eliminates the entire class of "permission denied" failures while maintaining a zero-trust security posture — no sudo passwords are ever handled by the application.

**Independent Test**: Can be fully tested by registering a new VM where the SSH user has standard sudo access (password-required), verifying that the provisioning engine runs automatically, and then confirming that service restart, log streaming, and power commands execute without any password prompt or error.

**Acceptance Scenarios**:

1. **Given** an administrator registers a new VM with valid SSH credentials, **When** the registration completes, **Then** the platform automatically executes the provisioning script on the remote host via SSH, configuring scoped NOPASSWD rules.
2. **Given** provisioning has completed successfully, **When** the administrator triggers any privileged action (service start/stop/restart, log stream, reboot), **Then** the action executes immediately without any sudo password prompt.
3. **Given** the provisioning script is executing, **When** the sudoers syntax validation check fails, **Then** the provisioning is aborted, no changes are applied to the target VM, and the administrator sees a clear error indicating the provisioning failure.
4. **Given** a VM was registered before the V2 upgrade (already has passwordless sudo or manual configuration), **When** the administrator uses the platform, **Then** all existing operations continue to function without re-provisioning (backward compatible).
5. **Given** provisioning has completed, **When** the SSH user attempts to run a command NOT in the allowed list (e.g., `sudo rm -rf /`), **Then** the standard sudo password prompt is still required — only platform-specific commands are authorized.
6. **Given** a VM registration is in progress, **When** the provisioning script encounters a network interruption, **Then** the VM is still registered but flagged as "Provisioning Failed" with an option to retry.

---

### User Story 2 - Multi-Page Dashboard Architecture (Priority: P1)

As a system administrator, I want the platform to have a structured multi-page layout with dedicated pages for the inventory, individual VM details, audit history, and settings, so that I can navigate complex server infrastructure without being confined to a single-page modal-based interface.

**Why this priority**: The V1 single-page architecture with slide-out panels becomes unmanageable as the number of VMs and features grows. A proper routing architecture is essential to support the additional telemetry pages and provide a professional enterprise-grade navigation experience.

**Independent Test**: Can be fully tested by navigating between the Inventory listing, a specific VM's detail page, the Audit Log page, and a Settings page, verifying each renders independently with its own URL and that browser back/forward navigation works correctly.

**Acceptance Scenarios**:

1. **Given** the user is on the Inventory page, **When** they click on a specific VM, **Then** they are navigated to a dedicated full-page VM Details view with its own URL (e.g., `/vms/{id}`).
2. **Given** the user is on a VM Detail page, **When** they click the "Audit Log" link in the navigation sidebar, **Then** they are navigated to a global Audit Log page showing all recorded platform events.
3. **Given** the user is on any page, **When** they use the browser's back button, **Then** they are returned to the previous page with its state preserved.
4. **Given** the user has a direct URL to a VM detail page, **When** they paste it into a new browser tab, **Then** the page loads correctly with all data populated (deep linking).
5. **Given** the user is on the Inventory page, **When** they access the navigation sidebar, **Then** they see clearly labeled links to: Dashboard/Overview, Inventory, Audit Log, and Settings.

---

### User Story 3 - Advanced Network Telemetry (Priority: P2)

As a system administrator, I want to see real-time network traffic statistics including bandwidth utilization (bytes in/out per second), active connections count, and top network-consuming processes, so that I can diagnose network bottlenecks and identify unexpected traffic patterns.

**Why this priority**: V1 only captures CPU, RAM, and Disk. Network telemetry is the most commonly requested missing metric for infrastructure troubleshooting. It enables administrators to detect DDoS patterns, runaway downloads, and misconfigured services.

**Independent Test**: Can be fully tested by generating network traffic on a target VM (e.g., downloading a file) and verifying that the dashboard displays increasing bytes-per-second values and the active connection count reflects the open sockets.

**Acceptance Scenarios**:

1. **Given** a VM is being monitored, **When** the user views its detail page, **Then** they see a dedicated Network panel displaying live bytes received/sent per second as a time-series chart.
2. **Given** a VM has active network connections, **When** the user views the network panel, **Then** a count of active TCP connections is displayed and updates in real-time.
3. **Given** a VM experiences a sudden spike in network activity, **When** the metric refresh occurs, **Then** the chart visually reflects the spike within 10 seconds.

---

### User Story 4 - Process Explorer (Priority: P2)

As a system administrator, I want to view a live, sorted list of the top resource-consuming processes on a VM, so that I can identify runaway processes or resource hogs without opening a terminal and running `htop` manually.

**Why this priority**: Process-level visibility is the natural next step after system-level CPU/RAM metrics. It answers the critical question "what is consuming all the resources?" directly in the dashboard.

**Independent Test**: Can be fully tested by launching a known resource-intensive process on a target VM and verifying that it appears near the top of the process list in the dashboard within one refresh cycle.

**Acceptance Scenarios**:

1. **Given** a VM is being monitored, **When** the user navigates to the Processes tab, **Then** they see a table of the top 20 processes sorted by CPU usage by default.
2. **Given** the process list is displayed, **When** the user clicks on the "Memory" column header, **Then** the list re-sorts by memory consumption.
3. **Given** a new high-CPU process starts on the VM, **When** the next telemetry refresh occurs, **Then** the process appears in the list at its correct rank.

---

### Edge Cases

- What happens if the SSH user does not have passwordless sudo or root access? The provisioning fails immediately with a clear error message instructing the administrator to grant the SSH user passwordless sudo (or use root) before retrying.
- How does the system handle VMs running non-standard Linux distributions where `visudo`, `systemctl`, or `journalctl` are not available or are located in non-standard paths?
- What happens if an administrator manually removes the provisioning rules from `/etc/sudoers.d/`? The platform does not actively detect drift. When a privileged command subsequently fails with a permission error, the error is surfaced to the user alongside a "Re-provision" button to restore the rules.
- How does the provisioning engine handle re-provisioning if the script is updated in a new platform version?
- How does the process list handle zombie or kernel threads that may have unusual display characteristics?
- What happens when a user navigates to a VM detail page for a VM that has been deleted by another administrator?
- How does the system handle network telemetry collection on VMs that have multiple network interfaces? Traffic is aggregated across all non-loopback, non-virtual interfaces. Loopback (`lo`), Docker bridges (`docker0`), virtual Ethernet pairs (`veth*`), and bridge interfaces (`br-*`) are excluded from the aggregate.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Upon successful VM registration, the platform MUST automatically execute a provisioning step on the target VM that configures the SSH user with scoped, passwordless sudo access for platform-required commands only.
- **FR-002**: The provisioning process MUST create a dedicated sudoers configuration file on the target VM (in `/etc/sudoers.d/`) granting NOPASSWD access exclusively to `systemctl` and `journalctl` commands — no other privileged commands.
- **FR-003**: The provisioning process MUST validate the generated sudoers syntax before applying it, aborting without changes if validation fails, to prevent administrator lockouts.
- **FR-004**: The platform MUST NOT transmit, accept, store, or cache sudo passwords at any point in its frontend, backend API, or database. All privileged execution is pre-authorized on the target host.
- **FR-005**: The provisioning script MUST be embedded within the backend binary at build time, ensuring it is version-controlled and cannot be modified at runtime.
- **FR-006**: The platform MUST track the provisioning status of each VM (e.g., "Provisioned", "Provisioning Failed", "Not Provisioned") and display it in the VM inventory.
- **FR-007**: The platform MUST provide a manual "Re-provision" action for VMs that failed initial provisioning or need their configuration refreshed.
- **FR-008**: The platform MUST implement a multi-page navigation architecture with dedicated routes for: Inventory Overview, VM Detail View, Audit Log, and Settings.
- **FR-009**: The platform MUST support deep linking, meaning any page can be accessed directly via its URL without requiring navigation from the home page.
- **FR-010**: The platform MUST display real-time network throughput metrics (bytes received and sent per second) for each monitored VM as time-series visualizations.
- **FR-011**: The platform MUST display the count of active TCP connections for each monitored VM.
- **FR-012**: The platform MUST display a sortable, auto-refreshing table of the top 20 processes on a VM ranked by resource consumption (CPU or memory).
- **FR-013**: The platform MUST render a persistent navigation sidebar that provides access to all top-level pages from any location in the application.
- **FR-014**: The platform MUST maintain backward compatibility with VMs registered under V1 that already have passwordless sudo configured.
- **FR-015**: The platform MUST log every provisioning attempt (success or failure) in the audit trail, including the target VM, the initiating user, and the outcome.

### Key Entities

- **Provisioning Record**: A persistent record tracking the provisioning state of a target VM. Contains the VM ID, provisioning status (Provisioned, Failed, Pending, Not Provisioned), timestamp of last provisioning attempt, and any error message from a failed attempt.
- **Network Metric**: A time-series data point capturing bytes received, bytes sent, and active TCP connection count for a specific VM at a specific moment.
- **Process Snapshot**: A point-in-time capture of the top processes running on a VM, including process ID, name, CPU percentage, memory percentage, and user owner.
- **Navigation Route**: A named, addressable page within the application that supports direct URL access and browser history integration.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Newly registered VMs are fully provisioned (sudo rules applied and validated) within 30 seconds of registration completing.
- **SC-002**: After provisioning, all privileged platform actions (service management, log streaming, power commands) execute without any password prompt and complete within the existing V1 performance thresholds.
- **SC-003**: Network telemetry metrics (throughput and connections) refresh at least once every 5 seconds, consistent with existing CPU/RAM polling.
- **SC-004**: The process list displays the top 20 processes and refreshes within 5 seconds of the previous snapshot.
- **SC-005**: Page navigation between any two pages in the application completes in under 500 milliseconds with no full-page reloads.
- **SC-006**: The system simultaneously supports the new telemetry collection (network + processes) alongside existing metrics for up to 50 target VMs without degrading response times.
- **SC-007**: A failed provisioning attempt is reported to the administrator within 5 seconds of the failure, with a clear error message and a retry option.

## Assumptions

- The SSH user account used for VM registration MUST already have passwordless sudo access or be the root user. If the SSH user requires a password for sudo, provisioning will fail and the administrator must grant passwordless sudo or use root before retrying. No sudo passwords are ever accepted by the platform.
- Target VMs have `visudo` installed and accessible, as it is part of the standard `sudo` package on all major Linux distributions.
- Target VMs use systemd-based service management (`systemctl`, `journalctl`). Non-systemd distributions are out of scope.
- Network telemetry aggregates traffic across all non-loopback, non-virtual interfaces (excluding `lo`, `docker0`, `veth*`, `br-*`). Interface names vary across distributions but the platform reads `/proc/net/dev` which lists all interfaces regardless of naming convention.
- The existing V1 authentication, RBAC, and audit logging infrastructure is reused and extended for V2 features.
- The process list is read-only; remote process management (kill, nice) is out of scope for this version.
- The Settings page in V2 is limited to basic platform configuration; full settings are deferred to a future version.
- The provisioning script creates rules scoped to the specific SSH username used during registration. If the username changes, re-provisioning is required.
