SET statement_timeout = 0;

--bun:split
DROP FUNCTION IF EXISTS insert_welcome_message_v2(BYTEA, BYTEA, BYTEA, BYTEA, SMALLINT);

--bun:split
ALTER TABLE welcome_messages
	DROP COLUMN IF EXISTS wrapper_algorithm;

