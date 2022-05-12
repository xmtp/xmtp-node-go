SET statement_timeout = 0;

--bun:split

DROP INDEX message_sort_idx;

--bun:split

DROP INDEX message_senderTimestamp_idx;

--bun:split

DROP INDEX message_contentTopic_idx;

--bun:split

DROP INDEX message_pubsubTopic_idx;