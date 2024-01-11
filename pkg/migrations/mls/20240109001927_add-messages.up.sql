SET statement_timeout = 0;

--bun:split

CREATE TABLE group_messages (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    group_id BYTEA NOT NULL,
    data BYTEA
);

--bun:split

CREATE INDEX idx_group_messages_group_id_created_at ON group_messages(group_id, created_at);

--bun:split

CREATE TABLE welcome_messages (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    installation_id BYTEA NOT NULL,
    data BYTEA
);

--bun:split

CREATE INDEX idx_welcome_messages_installation_id_created_at ON welcome_messages(installation_id, created_at);
