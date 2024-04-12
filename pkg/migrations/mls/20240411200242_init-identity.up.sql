SET statement_timeout = 0;

--bun:split

CREATE TABLE inbox_log (
    sequence_id BIGSERIAL PRIMARY KEY,
    inbox_id TEXT NOT NULL,
    commit_status SMALLINT NOT NULL,
    server_timestamp_ns BIGINT NOT NULL,
    identity_update_proto BYTEA NOT NULL
);

--bun:split

CREATE INDEX idx_inbox_log_inbox_id_commit_status ON inbox_log(inbox_id, commit_status);

--bun:split

CREATE TABLE address_log (
    inbox_log_sequence_id BIGINT PRIMARY KEY,
    address TEXT NOT NULL,
    identity_update_proto BYTEA NOT NULL
);

--bun:split

CREATE INDEX idx_address_log_address ON address_log(address);

--bun:split

CREATE TABLE address_lookup_cache (
    address TEXT PRIMARY KEY,
    inbox_id TEXT,
);
