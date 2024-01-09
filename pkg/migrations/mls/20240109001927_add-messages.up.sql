SET statement_timeout = 0;

--bun:split

CREATE TABLE messages (
    tid VARCHAR(300) PRIMARY KEY,
    topic VARCHAR(200) NOT NULL,
    created_at BIGINT NOT NULL,
    content BYTEA
);

--bun:split

CREATE INDEX idx_messages_tid ON messages(tid);
