ALTER TABLE group_messages DROP COLUMN is_commit IF EXISTS;

DROP INDEX IF EXISTS idx_groups_id_is_commit;