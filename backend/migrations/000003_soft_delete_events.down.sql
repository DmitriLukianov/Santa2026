DROP INDEX IF EXISTS idx_events_deleted_at;
ALTER TABLE events DROP COLUMN IF EXISTS deleted_at;
