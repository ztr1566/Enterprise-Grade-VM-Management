CREATE TABLE IF NOT EXISTS metrics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    vm_id TEXT NOT NULL,
    timestamp INTEGER NOT NULL,
    cpu_usage REAL,
    memory_used INTEGER,
    memory_total INTEGER,
    disk_usage REAL,
    network_tx INTEGER,
    network_rx INTEGER
);
CREATE INDEX IF NOT EXISTS idx_metrics_vm_ts ON metrics(vm_id, timestamp);

CREATE TABLE IF NOT EXISTS logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    vm_id TEXT NOT NULL,
    timestamp INTEGER NOT NULL,
    severity TEXT,
    component TEXT,
    message TEXT
);
CREATE INDEX IF NOT EXISTS idx_logs_vm_ts ON logs(vm_id, timestamp);
