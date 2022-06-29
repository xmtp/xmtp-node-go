SET statement_timeout = 0;

--bun:split

CREATE INDEX CONCURRENTLY message_index ON public.message (pubsubTopic, senderTimestamp, id);
