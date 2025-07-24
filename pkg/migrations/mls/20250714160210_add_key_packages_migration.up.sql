SET statement_timeout = 0;

ALTER TABLE installations
	ADD COLUMN is_appended BOOLEAN DEFAULT NULL;

--bun:split
CREATE TABLE key_packages(
	sequence_id BIGSERIAL PRIMARY KEY,
	installation_id BYTEA NOT NULL,
	key_package BYTEA NOT NULL,
	created_at BIGINT NOT NULL,
	UNIQUE (installation_id, key_package)
);

--bun:split
CREATE INDEX idx_key_packages_installation_id ON key_packages(installation_id);

