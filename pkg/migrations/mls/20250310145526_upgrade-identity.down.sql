SET
    statement_timeout = 0;

DROP INDEX idx_address_log_identifier_inbox_id;

ALTER TABLE address_log
DROP COLUMN identifier_kind
RENAME COLUMN identifier TO address;

CREATE INDEX idx_address_log_address_inbox_id ON address_log (address, inbox_id);
