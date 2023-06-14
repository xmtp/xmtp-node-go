SET statement_timeout = 0;

--bun:split

CREATE INDEX CONCURRENTLY message_ctopic_sts_id_idx ON public.message (contentTopic, senderTimestamp, id);
