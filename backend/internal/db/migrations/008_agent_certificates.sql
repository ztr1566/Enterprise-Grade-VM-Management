-- Migration 008: agent_certificates for tracking and revocation
CREATE TABLE IF NOT EXISTS agent_certificates (
    serial_number TEXT PRIMARY KEY,
    machine_id TEXT NOT NULL,
    issued_at DATETIME NOT NULL,
    expires_at DATETIME NOT NULL,
    revoked_at DATETIME,
    revocation_reason TEXT
);

CREATE INDEX IF NOT EXISTS idx_agent_certificates_machine_id ON agent_certificates(machine_id);
