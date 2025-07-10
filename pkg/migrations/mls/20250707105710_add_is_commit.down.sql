ALTER TABLE group_messages DROP COLUMN IF EXISTS is_commit;

DROP FUNCTION IF EXISTS insert_group_message_with_is_commit;