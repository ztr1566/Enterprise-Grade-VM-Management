-- Add last_seen_at to track agent heartbeats
ALTER TABLE vms ADD COLUMN last_seen_at DATETIME;
