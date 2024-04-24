-- name: GetAllInboxLogs :many
SELECT * FROM inbox_log
WHERE inbox_id = $1
ORDER BY sequence_id ASC FOR UPDATE;

-- name: GetInboxLogs :many
WITH b (inbox_id, sequence_id) AS (
    SELECT * FROM unnest(@filters::record[]) AS (inbox_id TEXT, sequence_id BIGINT)
)
SELECT a.* FROM inbox_log AS a
JOIN b ON a.inbox_id = b.inbox_id AND a.sequence_id > b.sequence_id
ORDER BY a.sequence_id ASC;