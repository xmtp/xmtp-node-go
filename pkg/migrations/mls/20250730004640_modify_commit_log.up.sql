SET statement_timeout = 0;

-- TODO(rich): Drop the old commit log table and function once servers have been deployed
--bun:split
CREATE TABLE IF NOT EXISTS commit_log_v2(
	id BIGSERIAL PRIMARY KEY,
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
	group_id BYTEA NOT NULL,
	serialized_entry BYTEA NOT NULL,
	serialized_signature BYTEA NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_commit_log_v2_group_id_id ON commit_log_v2(group_id, id);

--bun:split
CREATE FUNCTION insert_commit_log_v2(group_id BYTEA, serialized_entry BYTEA, serialized_signature BYTEA)
	RETURNS SETOF commit_log_v2
	AS $$
BEGIN
	-- Ensures that the generated sequence ID matches the insertion order
	-- Only released at the end of the enclosing transaction - beware if called within a long transaction
	PERFORM
		pg_advisory_xact_lock(hashtext('commit_log_sequence'), hashtext(encode(group_id, 'hex')));
	RETURN QUERY INSERT INTO commit_log_v2(group_id, serialized_entry, serialized_signature)
		VALUES(group_id, serialized_entry, serialized_signature)
	RETURNING
		*;
END;
$$
LANGUAGE plpgsql;

