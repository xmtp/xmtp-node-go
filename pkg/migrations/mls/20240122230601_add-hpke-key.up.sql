SET statement_timeout = 0;

ALTER TABLE welcome_messages
ADD COLUMN hpke_public_key BYTEA;
