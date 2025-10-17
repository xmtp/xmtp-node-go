SET
    statement_timeout = 0;

--bun:split
-- Add a column to discriminate between regular welcome messages and welcome pointers
ALTER TABLE welcome_messages
ADD COLUMN message_type SMALLINT NOT NULL DEFAULT 0;

--bun:split
-- Allow null hpke public key for welcome pointer messages
ALTER TABLE welcome_messages
ALTER COLUMN hpke_public_key DROP NOT NULL;

--bun:split
-- Create a new function to insert welcome pointer messages
CREATE OR REPLACE FUNCTION insert_welcome_pointer_message_v1(
        installation_key BYTEA,
        welcome_pointer_data BYTEA,
        installation_key_data_hash BYTEA,
        hpke_public_key BYTEA,
        wrapper_algorithm SMALLINT
    )
	RETURNS SETOF welcome_messages
	AS $$
BEGIN
	PERFORM
		pg_advisory_xact_lock(hashtext('welcome_messages_sequence'), hashtext(encode(installation_key, 'hex')));
	RETURN QUERY INSERT INTO welcome_messages(
		installation_key,
		data,
		installation_key_data_hash,
		hpke_public_key,
		wrapper_algorithm,
		message_type
	)
	VALUES(
		installation_key,
		welcome_pointer_data,
		installation_key_data_hash,
		hpke_public_key,
		wrapper_algorithm,
		1
	)
	RETURNING
		*;
END;
$$
LANGUAGE plpgsql;


