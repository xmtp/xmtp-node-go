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
    not_consumed BOOLEAN DEFAULT TRUE NOT NULL,
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
CREATE INDEX idx_key_packages_installation_id_not_consumed_is_last_resort_created_at ON key_packages(
    installation_id,
    not_consumed,
    is_last_resort,
    created_at
);

--bun:split
CREATE INDEX idx_key_packages_is_last_resort_id ON key_packages(is_last_resort, id);