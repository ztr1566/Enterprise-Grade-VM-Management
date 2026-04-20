# Quickstart: Hardened Agent

## 1. Generating an OTT
To provision a new agent, the administrator generates a One-Time-Token on the backend:
```bash
./backend provision-token --machine-id "vm-prod-001"
# Outputs: Token: <SECRET_OTT> (Valid for 10 minutes)
```

## 2. Starting the Agent
The agent binary expects the OTT as an environment variable or flag during its first run.
```bash
export AGENT_OTT="<SECRET_OTT>"
./agent -server 10.0.0.5:50051 -dir /var/lib/vm-agent/
```

- The agent generates an ECDSA P-256 key pair (`/var/lib/vm-agent/agent.key`).
- It submits the CSR with the OTT.
- The backend issues the 30-day certificate to `/var/lib/vm-agent/agent.crt`.

## 3. Operations & Recovery
Once enrolled, the agent runs continuously:
- **Resource Governor**: Keeps usage under 100MB RAM and 10% CPU.
- **WAL Recovery**: If the VM crashes or loses power, the agent truncates trailing corruption in the WAL and resumes streaming automatically on the next startup.
- **Monitoring**: Local metrics can be scraped: `curl http://127.0.0.1:9090/health`

## 4. Revocation
To immediately stop telemetry from a compromised agent:
```bash
./backend revoke-agent --machine-id "vm-prod-001"
```
The agent's mTLS connections will be rejected immediately upon the next reconnection.
