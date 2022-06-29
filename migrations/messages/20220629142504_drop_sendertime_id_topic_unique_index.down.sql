SET statement_timeout = 0;

--bun:split

CREATE INDEX CONCURRENTLY messageIndex ON public.message (pubsubTopic, receiverTimestamp, id);
