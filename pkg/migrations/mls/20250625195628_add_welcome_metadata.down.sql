SET
    statement_timeout = 0;

--bun:split
ALTER TABLE welcome_messages
DROP COLUMN welcome_metadata;

--bun:split
ALTER TABLE welcome_messages
ADD COLUMN message_cursor BIGINT NOT NULL DEFAULT 0;

--bun:split
DROP FUNCTION IF EXISTS insert_welcome_message_v4;
