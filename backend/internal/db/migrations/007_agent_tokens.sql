-- Migration 007: agent_tokens for OTT authorization
CREATE TABLE IF NOT EXISTS agent_tokens (
    id TEXT PRIMARY KEY,
    token_hash TEXT NOT NULL,
    machine_id TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    expires_at DATETIME NOT NULL,
    used_at DATETIME
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_agent_tokens_token_hash ON agent_tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_agent_tokens_machine_id ON agent_tokens(machine_id);
