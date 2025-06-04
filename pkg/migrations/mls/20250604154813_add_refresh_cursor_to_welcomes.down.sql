SET
    statement_timeout = 0;

ALTER TABLE welcome_messages
DROP COLUMN group_refresh_state_cursor;
