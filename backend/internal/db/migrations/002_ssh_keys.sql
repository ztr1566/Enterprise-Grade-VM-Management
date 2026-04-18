-- Migration 002: SSH Key Vault
-- Safe to re-run (IF NOT EXISTS guards)

CREATE TABLE IF NOT EXISTS ssh_keys (
    id         TEXT PRIMARY KEY,
    name       TEXT UNIQUE NOT NULL,
    private_key TEXT NOT NULL,       -- AES-256-GCM encrypted, base64
    public_key  TEXT NOT NULL,       -- plain OpenSSH public key (safe to store)
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Optional: allow VMs to reference a stored key by ID instead of inline credential
ALTER TABLE vms ADD COLUMN key_id TEXT REFERENCES ssh_keys(id);
