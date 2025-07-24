SET statement_timeout = 0;

ALTER TABLE installations
	DROP COLUMN IF EXISTS is_appended;

DROP TABLE IF EXISTS key_packages;

DROP INDEX IF EXISTS idx_key_packages_installation_id;

