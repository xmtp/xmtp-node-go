SET statement_timeout = 0;

--bun:split

CREATE INDEX CONCURRENTLY message_sort_idx ON public.message (senderTimestamp, id);

--bun:split

CREATE INDEX CONCURRENTLY message_senderTimestamp_idx ON public.message (senderTimestamp);

--bun:split

CREATE INDEX CONCURRENTLY message_contentTopic_idx ON public.message (contentTopic);

--bun:split

CREATE INDEX CONCURRENTLY message_pubsubTopic_idx ON public.message (pubsubTopic);