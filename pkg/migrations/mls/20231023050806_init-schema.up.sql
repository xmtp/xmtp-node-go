SET statement_timeout = 0;

--bun:split
CREATE TABLE installations(
	id BYTEA PRIMARY KEY,
	wallet_address TEXT NOT NULL,
	created_at BIGINT NOT NULL,
	updated_at BIGINT NOT NULL,
	credential_identity BYTEA NOT NULL,
	revoked_at BIGINT,
	key_package BYTEA NOT NULL,
	expiration BIGINT NOT NULL
);

--bun:split
CREATE INDEX idx_installations_wallet_address ON installations(wallet_address);

--bun:split
CREATE INDEX idx_installations_created_at ON installations(created_at);

--bun:split
CREATE INDEX idx_installations_revoked_at ON installations(revoked_at);

