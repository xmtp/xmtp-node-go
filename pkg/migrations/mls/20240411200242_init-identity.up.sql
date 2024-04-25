SET statement_timeout = 0;

--bun:split

CREATE TABLE inbox_log (
    sequence_id BIGSERIAL PRIMARY KEY,
    inbox_id TEXT NOT NULL,
    server_timestamp_ns BIGINT NOT NULL,
    identity_update_proto BYTEA NOT NULL
);

--bun:split

CREATE INDEX idx_inbox_log_inbox_id ON inbox_log(inbox_id);

--bun:split

CREATE TABLE address_log (
    address TEXT NOT NULL,
    inbox_id TEXT NOT NULL,
    association_sequence_id BIGINT,
    revocation_sequence_id BIGINT
);

--bun:split

CREATE INDEX idx_address_log_address_inbox_id ON address_log(address, inbox_id);