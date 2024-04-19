SET statement_timeout = 0;

--bun:split

CREATE INDEX CONCURRENTLY IF NOT EXISTS message_pubsubTopic_idx ON public.message (pubsubTopic);
