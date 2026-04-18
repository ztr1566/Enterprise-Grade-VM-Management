-- Migration 004: VM Access control table for multi-user provisioning
CREATE TABLE IF NOT EXISTS vm_access (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    vm_id TEXT NOT NULL,
    username TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (vm_id) REFERENCES vms(id) ON DELETE CASCADE,
    UNIQUE(vm_id, username)
);
