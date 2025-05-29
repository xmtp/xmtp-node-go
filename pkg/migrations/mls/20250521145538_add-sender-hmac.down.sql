SET statement_timeout = 0;

--bun:split
ALTER TABLE group_messages
	DROP COLUMN IF EXISTS sender_hmac;

--bun:split
ALTER TABLE should_push
	DROP COLUMN IF EXISTS should_push;

--bun:split
CREATE OR REPLACE FUNCTION insert_group_message(group_id BYTEA, data BYTEA, group_id_data_hash BYTEA)
	RETURNS SETOF group_messages
	AS $$
BEGIN
	-- Ensures that the generated sequence ID matches the insertion order
	-- Only released at the end of the enclosing transaction - beware if called within a long transaction
	PERFORM
		pg_advisory_xact_lock(hashtext('group_messages_sequence'), hashtext(encode(group_id, 'hex')));
	RETURN QUERY INSERT INTO group_messages(group_id, data, group_id_data_hash)
		VALUES(group_id, data, group_id_data_hash)
	RETURNING
		*;
END;
$$
LANGUAGE plpgsql;

