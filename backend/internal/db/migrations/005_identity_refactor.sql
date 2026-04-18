-- Migration 005: Identity Architecture Refactor
-- Rename username to management_username in vms table
ALTER TABLE vms RENAME COLUMN username TO management_username;
