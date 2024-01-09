SET statement_timeout = 0;

--bun:split

CREATE TABLE messages (
    topic VARCHAR(300) NOT NULL,
    tid VARCHAR(300) NOT NULL,
    created_at BIGINT NOT NULL,
    content BYTEA,
    CONSTRAINT idx_messages_topic_tid PRIMARY KEY (topic, tid)
);
