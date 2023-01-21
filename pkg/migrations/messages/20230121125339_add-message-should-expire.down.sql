SET statement_timeout = 0;

--bun:split

DROP INDEX message_recvts_shouldexpire_idx;

--bun:split

ALTER TABLE message DROP COLUMN should_expire;
