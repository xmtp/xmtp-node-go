SET statement_timeout = 0;

--bun:split
ALTER TABLE installations
	ADD COLUMN inbox_id BYTEA NOT NULL,
	ADD COLUMN expiration BIGINT NOT NULL;

