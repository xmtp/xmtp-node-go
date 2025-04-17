SET statement_timeout = 0;

--bun:split
-- Set this function back to the old version
CREATE OR REPLACE FUNCTION insert_welcome_message(installation_key BYTEA, data BYTEA, installation_key_data_hash BYTEA, hpke_public_key BYTEA)
	RETURNS SETOF welcome_messages
	AS $$
BEGIN
	PERFORM
		pg_advisory_xact_lock(hashtext('welcome_messages_sequence'), hashtext(encode(installation_key, 'hex')));
	RETURN QUERY INSERT INTO welcome_messages(installation_key, data, installation_key_data_hash, hpke_public_key)
		VALUES(installation_key, data, installation_key_data_hash, hpke_public_key)
	RETURNING
		*;
END;
$$
LANGUAGE plpgsql;

--bun:split

ALTER TABLE welcome_messages DROP COLUMN IF EXISTS wrapper_algorithm;