CREATE TABLE IF NOT EXISTS provisioning_records (
    vm_id           TEXT PRIMARY KEY REFERENCES vms(id) ON DELETE CASCADE,
    status          TEXT NOT NULL DEFAULT 'not_provisioned',
    last_attempt_at DATETIME,
    error_message   TEXT,
    provisioned_user TEXT,
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP
);
