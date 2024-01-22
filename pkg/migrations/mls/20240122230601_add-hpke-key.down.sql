SET statement_timeout = 0;

ALTER TABLE welcome_messages
REMOVE COLUMN IF EXISTS hpke_public_key BYTEA;
