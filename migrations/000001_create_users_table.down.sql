-- 000001_create_users_table.down.sql
-- Rolls back the users table creation.
-- Run with: make migrate-down

DROP INDEX IF EXISTS idx_users_email;
DROP TABLE IF EXISTS users;
