SET statement_timeout = 0;

--bun:split

CREATE TABLE IF NOT EXISTS inboxes(
    id BYTEA PRIMARY KEY,
    updated_at TIMESTAMP DEFAULT NOW()
);
