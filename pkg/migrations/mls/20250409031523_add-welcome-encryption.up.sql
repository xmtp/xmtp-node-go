SET statement_timeout = 0;

--bun:split
ALTER TABLE welcome_messages
	ADD COLUMN wrapper_algorithm SMALLINT NOT NULL DEFAULT 0;

--bun:split
CREATE OR REPLACE FUNCTION insert_welcome_message_v2(installation_key BYTEA, data BYTEA, installation_key_data_hash BYTEA, hpke_public_key BYTEA, wrapper_algorithm SMALLINT)
	RETURNS SETOF welcome_messages
	AS $$
BEGIN
	PERFORM
		pg_advisory_xact_lock(hashtext('welcome_messages_sequence'), hashtext(encode(installation_key, 'hex')));
	RETURN QUERY INSERT INTO welcome_messages(installation_key, data, installation_key_data_hash, hpke_public_key, wrapper_algorithm)
		VALUES(installation_key, data, installation_key_data_hash, hpke_public_key, wrapper_algorithm)
	RETURNING
		*;
END;
$$
LANGUAGE plpgsql;
