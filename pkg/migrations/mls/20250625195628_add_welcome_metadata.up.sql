SET
    statement_timeout = 0;

--bun:split
ALTER TABLE welcome_messages
DROP COLUMN message_cursor;

--bun:split
ALTER TABLE welcome_messages
ADD COLUMN message_metadata BYTEA NOT NULL DEFAULT ''::bytea;

--bun:split
CREATE OR REPLACE FUNCTION insert_welcome_message_v3(
        installation_key BYTEA,
        data BYTEA,
        installation_key_data_hash BYTEA,
        hpke_public_key BYTEA,
        wrapper_algorithm SMALLINT,
        message_metadta BYTEA
    )
	RETURNS SETOF welcome_messages
	AS $$
BEGIN
	PERFORM
		pg_advisory_xact_lock(hashtext('welcome_messages_sequence'), hashtext(encode(installation_key, 'hex')));
	RETURN QUERY INSERT INTO welcome_messages(installation_key, data, installation_key_data_hash, hpke_public_key, wrapper_algorithm, message_metadata)
	VALUES(installation_key, data, installation_key_data_hash, hpke_public_key, wrapper_algorithm, message_metadata)
	RETURNING
		*;
END;
$$
LANGUAGE plpgsql;
