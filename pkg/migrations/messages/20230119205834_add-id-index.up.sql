SET statement_timeout = 0;

--bun:split

CREATE INDEX CONCURRENTLY message_id_idx ON public.message (id);
