# VM Telemetry Agent Quickstart

## Backend CA & gRPC Setup
1. The backend automatically initializes an internal CA on startup if one does not exist.
2. The gRPC server listens on a dedicated secure port for Agent Identity (CSRs) and Telemetry Ingestion.

## Agent Provisioning
To deploy the agent to a new VM:
```bash
./scripts/provisioning/deploy_agent.sh <vm_ip> <vm_id>
```
This script uses SSH *only* to transfer the compiled `agent` binary and configure its systemd service.

## Agent Bootstrap
Upon starting via systemd, the agent:
1. Generates a local private key (`/var/lib/vm-agent/client.key`).
2. Submits a CSR to the backend using a one-time bootstrap token or pre-shared VM identity.
3. Receives and stores the signed mTLS cert (`/var/lib/vm-agent/client.crt`).
4. Begins hardware sampling and log aggregation.
5. Buffers to the local 50MB disk-backed WAL and transmits in batches over mTLS via gRPC.