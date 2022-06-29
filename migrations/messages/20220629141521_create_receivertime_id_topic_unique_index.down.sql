SET statement_timeout = 0;

--bun:split

ALTER TABLE public.message DROP CONSTRAINT message_index;
