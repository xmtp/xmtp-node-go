SET statement_timeout = 0;

--bun:split
ALTER TABLE group_messages
	ADD COLUMN sender_hmac BYTEA;

--bun:split
ALTER TABLE group_messages
	ADD COLUMN should_push BOOLEAN;

--bun:split
CREATE OR REPLACE FUNCTION insert_group_message_v3(group_id BYTEA, data BYTEA, group_id_data_hash BYTEA, sender_hmac BYTEA, should_push BOOLEAN, is_commit BOOLEAN)
	RETURNS SETOF group_messages
	AS $$
BEGIN
	-- Ensures that the generated sequence ID matches the insertion order
	-- Only released at the end of the enclosing transaction - beware if called within a long transaction
	PERFORM
		pg_advisory_xact_lock(hashtext('group_messages_sequence'), hashtext(encode(group_id, 'hex')));
	RETURN QUERY INSERT INTO group_messages(group_id, data, group_id_data_hash, sender_hmac, should_push, is_commit)
		VALUES(group_id, data, group_id_data_hash, sender_hmac, should_push, is_commit)
	RETURNING
		*;
END;
$$
LANGUAGE plpgsql;

