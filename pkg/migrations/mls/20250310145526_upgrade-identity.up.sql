SET
    statement_timeout = 0;

DROP INDEX idx_address_log_address_inbox_id;

ALTER TABLE address_log
ADD COLUMN identifier_kind INT,
RENAME COLUMN address TO identifier;

-- Default all of the existing identifier_kinds to 1 (Ethereum)
UPDATE address_log
SET
    identifier_kind = 1;

ALTER TABLE address_log
ALTER COLUMN identifier_kind
SET
    NOT NULL;

CREATE INDEX idx_address_log_identifier_inbox_id ON address_log (identifier, identifier_kind, inbox_id);
