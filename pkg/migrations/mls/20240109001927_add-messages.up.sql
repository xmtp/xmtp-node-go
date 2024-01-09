SET statement_timeout = 0;

--bun:split

CREATE TABLE messages (
    topic VARCHAR(200) NOT NULL,
    tid VARCHAR(96) NOT NULL,
    created_at BIGINT NOT NULL,
    content BYTEA,
    CONSTRAINT idx_messages_topic_tid PRIMARY KEY (topic, tid)
);

--bun:split

CREATE INDEX idx_messages_topic_created_at ON messages(topic, created_at);
