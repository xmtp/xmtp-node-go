--bun:split
CREATE TABLE key_packages(
	sequence_id BIGSERIAL PRIMARY KEY,
	installation_id BYTEA NOT NULL,
	key_package BYTEA NOT NULL
);

--bun:split
CREATE TABLE key_packages_backfill_tracker(
	last_migrated_timestamp BIGINT PRIMARY KEY NOT NULL
);

--bun:split
INSERT INTO key_packages_backfill_tracker(last_migrated_timestamp)
	VALUES (0);

--bun:split
CREATE INDEX idx_key_packages_installation_id ON key_packages(installation_id);

