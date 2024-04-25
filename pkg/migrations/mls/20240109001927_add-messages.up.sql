SET statement_timeout = 0;

--bun:split
CREATE TABLE group_messages(
	id BIGSERIAL PRIMARY KEY,
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
	group_id BYTEA NOT NULL,
	data BYTEA NOT NULL,
	group_id_data_hash BYTEA NOT NULL
);

--bun:split
CREATE INDEX idx_group_messages_group_id_created_at ON group_messages(group_id, created_at);

--bun:split
CREATE UNIQUE INDEX idx_group_messages_group_id_data_hash ON group_messages(group_id_data_hash);

--bun:split
CREATE TABLE welcome_messages(
	id BIGSERIAL PRIMARY KEY,
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
	installation_key BYTEA NOT NULL,
	data BYTEA NOT NULL,
	installation_key_data_hash BYTEA NOT NULL
);

--bun:split
CREATE INDEX idx_welcome_messages_installation_key_created_at ON welcome_messages(installation_key, created_at);

--bun:split
CREATE UNIQUE INDEX idx_welcome_messages_group_key_data_hash ON welcome_messages(installation_key_data_hash);

