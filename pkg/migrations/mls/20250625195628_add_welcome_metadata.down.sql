SET
    statement_timeout = 0;

--bun:split
ALTER TABLE welcome_messages
DROP COLUMN welcome_metadata;

--bun:split
DROP FUNCTION IF EXISTS insert_welcome_message_v4;
