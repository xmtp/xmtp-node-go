ALTER TABLE group_messages DROP COLUMN is_commit IF EXISTS;

DROP FUNCTION IF EXISTS insert_group_message_with_is_commit;

DROP INDEX IF EXISTS idx_groups_id_is_commit;