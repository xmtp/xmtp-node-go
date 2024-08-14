SET statement_timeout = 0;

--bun:split
ALTER TABLE installations
	DROP COLUMN IF EXISTS inbox_id,
	DROP COLUMN IF EXISTS expiration;

