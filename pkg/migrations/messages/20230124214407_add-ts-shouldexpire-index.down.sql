SET statement_timeout = 0;

--bun:split

CREATE INDEX CONCURRENTLY message_recvts_shouldexpire_idx ON public.message (receiverTimestamp, should_expire);

--bun:split

DROP INDEX message_recvts_shouldexpiretrue_idx;
