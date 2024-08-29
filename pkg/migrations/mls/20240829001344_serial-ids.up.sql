CREATE FUNCTION insert_group_message(group_id BYTEA, data BYTEA, group_id_data_hash BYTEA)
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

CREATE FUNCTION insert_welcome_message(installation_key BYTEA, data BYTEA, installation_key_data_hash BYTEA, hpke_public_key BYTEA)
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

CREATE FUNCTION insert_inbox_log(inbox_id BYTEA, server_timestamp_ns BIGINT, identity_update_proto BYTEA)
	RETURNS SETOF inbox_log
	AS $$
BEGIN
	PERFORM
		pg_advisory_xact_lock(hashtext('inbox_log_sequence'), hashtext(encode(inbox_id, 'hex')));
	RETURN QUERY INSERT INTO inbox_log(inbox_id, server_timestamp_ns, identity_update_proto)
		VALUES(inbox_id, server_timestamp_ns, identity_update_proto)
	RETURNING
		*;
END;
$$
LANGUAGE plpgsql;

