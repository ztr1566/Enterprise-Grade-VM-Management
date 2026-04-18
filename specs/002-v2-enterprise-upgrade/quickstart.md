# Quickstart: V2 Enterprise Upgrade

**Feature**: V2 Enterprise Upgrade  
**Date**: 2026-04-18

## Scenario 1: Zero-Touch Provisioning on VM Registration

### Steps
1. Administrator logs in to the platform.
2. Clicks "Add Server" and fills in: Name, Host IP, Username, Auth Type (password or key), Credential.
3. Clicks "Save".
4. Backend creates the VM record in SQLite.
5. Backend immediately triggers the provisioning engine:
   a. Opens an SSH session to the new VM using the provided credentials.
   b. Uploads the embedded `setup_vm_sudoers.sh` script to `/tmp/`.
   c. Executes `sudo bash /tmp/vm-platform-provision.sh <username>`.
   d. The script creates `/etc/sudoers.d/vm-platform-<username>` with scoped NOPASSWD rules.
   e. The script runs `visudo -c` to validate syntax.
   f. If validation passes, the rules are applied. If it fails, the temp file is removed and no changes are made.
   g. The script cleans up the temp file.
6. Backend updates the `provisioning_records` table with the result.
7. Frontend displays the VM in the inventory with a provisioning status badge.

### Expected Outcome
- VM appears in inventory with status "Provisioned" (green badge).
- All privileged actions (service management, log streaming, power commands) work without sudo prompts.

### Failure Scenario
- If provisioning fails (e.g., user not in sudoers at all), the VM still appears in inventory but with a "Provisioning Failed" (red badge) and a "Retry" button.

---

## Scenario 2: Multi-Page Navigation

### Steps
1. User lands on `/inventory` after login.
2. Clicks on a VM row → navigated to `/vms/{id}`.
3. VM detail page shows tabs: Metrics, Services, Logs, Network, Processes.
4. User clicks "Audit Log" in the sidebar → navigated to `/audit`.
5. User clicks browser back button → returned to `/vms/{id}`.
6. User copies the URL `/vms/{id}` and pastes in a new tab → page loads correctly.

### Expected Outcome
- Each page has its own URL.
- Browser history works correctly.
- Sidebar is visible on every page.

---

## Scenario 3: Network Telemetry Observation

### Steps
1. User navigates to `/vms/{id}` and clicks the "Network" tab.
2. A real-time chart displays bytes received/sent per second.
3. A counter shows active TCP connections.
4. User starts a large file download on the target VM.
5. The chart spikes within 10 seconds.

### Expected Outcome
- Network metrics update every 5 seconds.
- The chart reflects real traffic patterns.

---

## Scenario 4: Process Explorer

### Steps
1. User navigates to `/vms/{id}` and clicks the "Processes" tab.
2. A table shows the top 20 processes sorted by CPU usage.
3. User clicks the "Memory" column header.
4. The table re-sorts by memory consumption.
5. User launches a stress test on the VM.
6. The stress process appears near the top on the next refresh.

### Expected Outcome
- Process list is accurate and sortable.
- Refreshes every 5 seconds.
