SET statement_timeout = 0;

--bun:split
CREATE TABLE key_packages(
	sequence_id BIGSERIAL PRIMARY KEY,
	installation_id BYTEA NOT NULL,
	key_package BYTEA NOT NULL,
	UNIQUE (installation_id, key_package)
);

--bun:split
CREATE INDEX idx_key_packages_installation_id ON key_packages(installation_id);

-- Migrate installations created or updated after 2025-04-16 00:00:00 UTC
-- Use advisory lock to ensure only one node performs the migration
--bun:split
DO $$
BEGIN
	IF pg_try_advisory_lock(hashtext('key_packages_migration_20250714')) THEN
		-- Lock both tables to prevent any writes during migration
		LOCK TABLE installations IN SHARE MODE;
		LOCK TABLE key_packages IN EXCLUSIVE MODE;
		INSERT INTO key_packages(installation_id, key_package)
		SELECT
			id,
			key_package
		FROM
			installations
		WHERE
			updated_at >= 1744761600000000000;
		PERFORM
			pg_advisory_unlock(hashtext('key_packages_migration_20250714'));
	END IF;
END
$$;

