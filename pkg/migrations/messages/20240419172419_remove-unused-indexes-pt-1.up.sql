SET statement_timeout = 0;

-- bun:split
DROP INDEX IF EXISTS message_pubsubtopic_idx;

-- bun:split
DROP INDEX IF EXISTS message_sort_idx;

-- bun:split
DROP INDEX IF EXISTS message_recvts_shouldexpire_idx;

-- bun:split
-- Replace recvts_shouldexpire_idx with an index just on receiver timestamp
-- This index is used in our data pipelines to power the data warehouse for high level analytics on network growth
CREATE INDEX CONCURRENTLY IF NOT EXISTS message_receivertimestamp_idx ON public.message(receiverTimestamp);

