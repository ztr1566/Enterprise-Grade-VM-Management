# Research: V2 Enterprise Upgrade — Zero-Touch Provisioning

**Feature**: V2 Enterprise Upgrade  
**Date**: 2026-04-18  
**Spec**: [spec.md](/home/ztr/Projects/new_project/specs/002-v2-enterprise-upgrade/spec.md)

## Research Tasks

### R1: Embedded Script Bundling with `//go:embed`

**Decision**: Use Go 1.16+ `//go:embed` directive to bundle `scripts/provisioning/setup_vm_sudoers.sh` into the `provisioner.go` binary at compile time.

**Rationale**: `//go:embed` makes the script immutable after build, ensuring version-controlled integrity. The script cannot be modified at runtime on the deployment host, satisfying architectural constraint #3 (tamper-resistant). No filesystem dependencies at runtime.

**Alternatives Considered**:
- **os.ReadFile at runtime**: Rejected — script could be modified on the server after deployment, breaking the tamper-resistance requirement.
- **Hardcoded string in Go**: Rejected — difficult to maintain, review, and test as a standalone bash script. Syntax highlighting and linting are lost.

### R2: Sudoers Provisioning via `/etc/sudoers.d/`

**Decision**: Create a drop-in file at `/etc/sudoers.d/vm-platform-<username>` containing scoped NOPASSWD rules for `systemctl` and `journalctl` only.

**Rationale**: Using `/etc/sudoers.d/` is the standard method for modular sudoers configuration on modern Linux. It avoids editing the main `/etc/sudoers` file, reducing the risk of breaking existing sudo configurations. The `visudo -c` check validates syntax before the file is moved into place.

**Alternatives Considered**:
- **Editing `/etc/sudoers` directly**: Rejected — high risk of corruption; difficult to manage idempotently.
- **Blanket `NOPASSWD: ALL`**: Rejected — violates architectural constraint #4 (principle of least privilege).

### R3: Provisioning Execution Strategy

**Decision**: Execute provisioning as a two-phase pipeline over the existing `crypto/ssh` infrastructure:
1. **Upload**: Write the embedded script to a temporary file on the remote host (`/tmp/vm-platform-provision.sh`).
2. **Execute**: Run the script with `sudo bash /tmp/vm-platform-provision.sh <username>`, passing the SSH username as an argument. The script handles sudoers creation, validation, and cleanup.

**Rationale**: Reuses the existing `internal/ssh` package (already proven for terminal, service management, and log streaming). No new dependencies required. The temporary file is cleaned up by the script after execution.

**Alternatives Considered**:
- **Piping script via stdin**: Rejected — some SSH server configurations restrict stdin for non-interactive sessions.
- **SCP/SFTP separate upload**: Rejected — adds complexity; `crypto/ssh` session already supports command execution with stdin/stdout.

### R4: Network Telemetry Collection Method

**Decision**: Collect network metrics by reading `/proc/net/dev` for byte counters and using `ss -tun state established | wc -l` for active TCP connection count. Compute rates (bytes/sec) by diffing consecutive samples on the backend.

**Rationale**: `/proc/net/dev` is universally available on Linux, requires no elevated privileges, and provides per-interface byte counters. `ss` is faster and more reliable than `netstat` on modern distributions.

**Alternatives Considered**:
- **`ifstat`/`nload`**: Rejected — not universally installed; would require provisioning additional packages.
- **`/sys/class/net/*/statistics/`**: Viable alternative but `/proc/net/dev` provides all interfaces in a single read.

### R5: Process List Collection Method

**Decision**: Use `ps aux --sort=-%cpu | head -21` to retrieve the top 20 processes sorted by CPU. Parse the output into structured data (PID, user, CPU%, MEM%, command).

**Rationale**: `ps` is universally available, requires no elevated privileges, and provides all needed fields in a single command. The `--sort` flag avoids client-side sorting overhead.

**Alternatives Considered**:
- **`top -bn1`**: Rejected — output format varies significantly across distributions; harder to parse reliably.
- **Reading `/proc/[pid]/stat` directly**: Rejected — requires complex parsing and multiple filesystem reads per process.

### R6: Multi-Page Frontend Architecture

**Decision**: Transition from the current single-page modal architecture to React Router v6 with the following route structure:
- `/` → Dashboard/Overview  
- `/inventory` → VM Inventory list  
- `/vms/:id` → VM Detail (full-page, tabbed: Metrics, Services, Logs, Processes, Network)  
- `/audit` → Audit Log  
- `/settings` → Settings  

**Rationale**: React Router v6 is already compatible with the Vite build system in use. Each route renders as an independent page with deep-linking support. The persistent sidebar provides global navigation.

**Alternatives Considered**:
- **Keeping modal architecture**: Rejected — does not scale for the additional telemetry tabs and violates spec US2.
- **Next.js migration**: Rejected — unnecessary complexity for a single-backend SPA; would require re-architecting the build pipeline.
