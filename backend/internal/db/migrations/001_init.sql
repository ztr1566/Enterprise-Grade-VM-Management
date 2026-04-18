CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS vms (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    host TEXT NOT NULL,
    username TEXT NOT NULL,
    auth_type TEXT NOT NULL,
    credential TEXT NOT NULL, -- AES encrypted
    tags TEXT, -- JSON array of strings
    status TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS audit_log_entries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    event_type TEXT NOT NULL, -- e.g., "credential_access", "vm_creation"
    details TEXT NOT NULL,    -- JSON or text details
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
);
