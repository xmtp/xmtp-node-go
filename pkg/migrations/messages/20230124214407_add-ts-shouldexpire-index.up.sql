SET statement_timeout = 0;

--bun:split

CREATE INDEX CONCURRENTLY message_recvts_shouldexpiretrue_idx ON public.message (receiverTimestamp) WHERE should_expire IS TRUE;

--bun:split

DROP INDEX message_recvts_shouldexpire_idx;
