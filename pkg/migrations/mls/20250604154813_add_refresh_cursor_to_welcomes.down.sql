SET
    statement_timeout = 0;

--bun:split
ALTER TABLE welcome_messages
DROP COLUMN message_cursor;
