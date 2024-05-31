SET statement_timeout = 0;

--bun:split
DROP TABLE IF EXISTS installations, group_messages, welcome_messages, inbox_log, address_log;

--bun:split
DROP TYPE IF EXISTS inbox_filter;

