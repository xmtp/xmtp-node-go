SET
    statement_timeout = 0;

--bun:split
-- Drop the new functions
DROP FUNCTION IF EXISTS insert_welcome_pointer_message_v1(BYTEA, BYTEA, BYTEA, BYTEA, SMALLINT);

--bun:split
ALTER TABLE welcome_messages
    DROP COLUMN IF EXISTS message_type;

--bun:split
ALTER TABLE welcome_messages
    ALTER COLUMN hpke_public_key SET NOT NULL;
