SET statement_timeout = 0;

--bun:split

CREATE TYPE inbox_filter AS (
    inbox_id TEXT,
    sequence_id BIGINT
);
