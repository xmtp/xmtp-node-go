ALTER TABLE group_messages ADD COLUMN is_commit BOOLEAN DEFAULT NULL;

CREATE OR REPLACE FUNCTION insert_group_message_with_is_commit(
	group_id BYTEA,
	data BYTEA,
	group_id_data_hash BYTEA,
	is_commit BOOLEAN DEFAULT NULL
)
RETURNS SETOF group_messages
AS $$
BEGIN
	-- Ensures that the generated sequence ID matches the insertion order
	-- Only released at the end of the enclosing transaction - beware if called within a long transaction
	PERFORM
pg_advisory_xact_lock(hashtext('group_messages_sequence'), hashtext(encode(group_id, 'hex')));

RETURN QUERY
    INSERT INTO group_messages(group_id, data, group_id_data_hash, is_commit)
		VALUES(group_id, data, group_id_data_hash, is_commit)
	RETURNING
		*;
END;
$$
LANGUAGE plpgsql;