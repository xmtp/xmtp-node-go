SET
    statement_timeout = 0;

--bun:split
ALTER TABLE
    authz_addresses DROP COLUMN comment;