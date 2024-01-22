SET statement_timeout = 0;

ALTER TABLE welcome_messages
DROP COLUMN IF EXISTS hpke_public_key BYTEA;
