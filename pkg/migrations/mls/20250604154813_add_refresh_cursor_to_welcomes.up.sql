SET
    statement_timeout = 0;

ALTER TABLE welcome_messages
ADD COLUMN group_refresh_state_cursor BIGINT NOT NULL DEFAULT 0;
