SET statement_timeout = 0;

--bun:split

ALTER TABLE message ADD column should_expire BOOLEAN DEFAULT FALSE;

--bun:split

CREATE INDEX CONCURRENTLY message_recvts_shouldexpire_idx ON public.message (receiverTimestamp, should_expire);
