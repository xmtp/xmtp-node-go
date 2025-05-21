SET statement_timeout = 0;

--bun:split
DROP FUNCTION IF EXISTS insert_group_message_v3(BYTEA, BYTEA, BYTEA, BYTEA, BOOLEAN, BOOLEAN);

--bun:split
ALTER TABLE group_messages
	DROP COLUMN IF EXISTS sender_hmac;

--bun:split
ALTER TABLE group_messages
	DROP COLUMN IF EXISTS should_push;

