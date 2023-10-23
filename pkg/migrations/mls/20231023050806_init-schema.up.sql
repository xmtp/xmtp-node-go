SET
    statement_timeout = 0;

--bun:split
CREATE TABLE installations (
    id TEXT PRIMARY KEY,
    wallet_address TEXT NOT NULL,
    created_at BIGINT NOT NULL,
    revoked_at BIGINT
);

--bun:split
CREATE TABLE key_packages (
    id TEXT PRIMARY KEY,
    installation_id TEXT NOT NULL,
    created_at BIGINT NOT NULL,
    consumed_at BIGINT,
    is_last_resort BOOLEAN NOT NULL,
    data BYTEA NOT NULL,
    -- Add a foreign key constraint to ensure key packages cannot be added for unregistered installations
    CONSTRAINT fk_installation_id FOREIGN KEY (installation_id) REFERENCES installations (id)
);

--bun:split
CREATE INDEX idx_installations_wallet_address ON installations(wallet_address);

--bun:split
CREATE INDEX idx_installations_created_at ON installations(created_at);

--bun:split
CREATE INDEX idx_installations_revoked_at ON installations(revoked_at);

--bun:split
-- Adding indexes for the key_packages table
CREATE INDEX idx_key_packages_installation_id ON key_packages(installation_id);

--bun:split
CREATE INDEX idx_key_packages_created_at ON key_packages(created_at);

--bun:split
CREATE INDEX idx_key_packages_consumed_at ON key_packages(consumed_at);

--bun:split
CREATE INDEX idx_key_packages_is_last_resort ON key_packages(is_last_resort);