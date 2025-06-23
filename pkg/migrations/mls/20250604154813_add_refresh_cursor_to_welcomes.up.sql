SET
    statement_timeout = 0;

--bun:split
ALTER TABLE welcome_messages
ADD COLUMN message_cursor BIGINT NOT NULL DEFAULT 0;

--bun:split
CREATE OR REPLACE FUNCTION insert_welcome_message_v3(
        installation_key BYTEA,
        data BYTEA,
        installation_key_data_hash BYTEA,
        hpke_public_key BYTEA,
        wrapper_algorithm SMALLINT,
        message_cursor BIGINT
    )
	RETURNS SETOF welcome_messages
	AS $$
BEGIN
	PERFORM
		pg_advisory_xact_lock(hashtext('welcome_messages_sequence'), hashtext(encode(installation_key, 'hex')));
	RETURN QUERY INSERT INTO welcome_messages(installation_key, data, installation_key_data_hash, hpke_public_key, wrapper_algorithm, message_cursor)
		VALUES(installation_key, data, installation_key_data_hash, hpke_public_key, wrapper_algorithm, message_cursor)
	RETURNING
		*;
END;
$$
LANGUAGE plpgsql;
