SET statement_timeout = 0;

--bun:split
CREATE TABLE IF NOT EXISTS commit_log(
	id BIGSERIAL PRIMARY KEY,
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
	group_id BYTEA NOT NULL,
	encrypted_entry BYTEA NOT NULL
);

CREATE FUNCTION insert_commit_log(group_id BYTEA, encrypted_entry BYTEA)
	RETURNS SETOF commit_log
	AS $$
BEGIN
	-- Ensures that the generated sequence ID matches the insertion order
	-- Only released at the end of the enclosing transaction - beware if called within a long transaction
	PERFORM
		pg_advisory_xact_lock(hashtext('commit_log_sequence'), hashtext(encode(group_id, 'hex')));
	RETURN QUERY INSERT INTO commit_log(group_id, encrypted_entry)
		VALUES(group_id, encrypted_entry)
	RETURNING
		*;
END;
$$
LANGUAGE plpgsql;

