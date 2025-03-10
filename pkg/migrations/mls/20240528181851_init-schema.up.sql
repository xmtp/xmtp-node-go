SET
    statement_timeout = 0;

--bun:split
CREATE TABLE installations (
    id BYTEA PRIMARY KEY,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    inbox_id BYTEA NOT NULL,
    key_package BYTEA NOT NULL,
    expiration BIGINT NOT NULL
);

--bun:split
CREATE TABLE group_messages (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW (),
    group_id BYTEA NOT NULL,
    data BYTEA NOT NULL,
    group_id_data_hash BYTEA NOT NULL
);

--bun:split
CREATE INDEX idx_group_messages_group_id_id ON group_messages (group_id, id);

--bun:split
CREATE UNIQUE INDEX idx_group_messages_group_id_data_hash ON group_messages (group_id_data_hash);

--bun:split
CREATE TABLE welcome_messages (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW (),
    installation_key BYTEA NOT NULL,
    data BYTEA NOT NULL,
    hpke_public_key BYTEA NOT NULL,
    installation_key_data_hash BYTEA NOT NULL
);

--bun:split
CREATE INDEX idx_welcome_messages_installation_key_id ON welcome_messages (installation_key, id);

--bun:split
CREATE UNIQUE INDEX idx_welcome_messages_group_key_data_hash ON welcome_messages (installation_key_data_hash);

--bun:split
CREATE TABLE inbox_log (
    sequence_id BIGSERIAL PRIMARY KEY,
    inbox_id BYTEA NOT NULL,
    server_timestamp_ns BIGINT NOT NULL,
    identity_update_proto BYTEA NOT NULL
);

--bun:split
CREATE INDEX idx_inbox_log_inbox_id_sequence_id ON inbox_log (inbox_id, sequence_id);

--bun:split
CREATE TABLE address_log (
    address TEXT NOT NULL,
    inbox_id BYTEA NOT NULL,
    association_sequence_id BIGINT,
    revocation_sequence_id BIGINT
);

--bun:split
CREATE INDEX idx_address_log_address_inbox_id ON address_log (address, inbox_id);

--bun:split
CREATE TYPE inbox_filter AS (
    inbox_id TEXT, -- Because this is serialized as JSON, we can't use a BYTEA type
    sequence_id BIGINT
);
