CREATE TABLE message (
    id BYTEA,
    receiverTimestamp BIGINT NOT NULL,
    senderTimestamp BIGINT NOT NULL,
    contentTopic TEXT NOT NULL,
    pubsubTopic TEXT NOT NULL,
    payload BYTEA,
    version INTEGER NOT NULL DEFAULT 0,
    CONSTRAINT messageIndex PRIMARY KEY (senderTimestamp, id, pubsubTopic)
);