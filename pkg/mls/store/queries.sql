-- name: GetAllInboxLogs :many
SELECT *
FROM inbox_log
WHERE inbox_id = $1
ORDER BY sequence_id ASC FOR
UPDATE;
-- name: GetInboxLogFiltered :many
SELECT a.*
FROM inbox_log AS a
    JOIN (
        SELECT *
        FROM json_populate_recordset(null::inbox_filter, sqlc.arg(filters)) as b(inbox_id, sequence_id)
    ) as b on b.inbox_id = a.inbox_id
    AND a.sequence_id > b.sequence_id
ORDER BY a.sequence_id ASC;
-- name: GetAddressLogs :many
SELECT a.address,
    a.inbox_id,
    a.association_sequence_id
FROM address_log a
    INNER JOIN (
        SELECT address,
            MAX(association_sequence_id) AS max_association_sequence_id
        FROM address_log
        WHERE address = ANY (@addresses::text [])
            AND revocation_sequence_id IS NULL
        GROUP BY address
    ) b ON a.address = b.address
    AND a.association_sequence_id = b.max_association_sequence_id;
-- name: InsertInboxLog :one
INSERT INTO inbox_log (
        inbox_id,
        server_timestamp_ns,
        identity_update_proto
    )
VALUES ($1, $2, $3)
RETURNING sequence_id;
-- name: CreateInstallation :exec
INSERT INTO installations (
        id,
        wallet_address,
        created_at,
        updated_at,
        credential_identity,
        key_package,
        expiration
    )
VALUES ($1, $2, $3, $3, $4, $5, $6);
-- name: UpdateKeyPackage :execrows
UPDATE installations
SET key_package = @key_package,
    updated_at = @updated_at,
    expiration = @expiration
WHERE id = @id;
-- name: FetchKeyPackages :many
SELECT id,
    key_package
FROM installations
WHERE id = ANY (sqlc.arg(installation_ids)::bytea []);
-- name: GetIdentityUpdates :many
SELECT *
FROM installations
WHERE wallet_address = ANY (@wallet_addresses::text [])
    AND (
        created_at > @start_time
        OR revoked_at > @start_time
    )
ORDER BY created_at ASC;
-- name: RevokeInstallation :exec
UPDATE installations
SET revoked_at = @revoked_at
WHERE id = @installation_id
    AND revoked_at IS NULL;
-- name: InsertGroupMessage :one
INSERT INTO group_messages (group_id, data, group_id_data_hash)
VALUES ($1, $2, $3)
RETURNING *;
-- name: InsertWelcomeMessage :one
INSERT INTO welcome_messages (
        installation_key,
        data,
        installation_key_data_hash,
        hpke_public_key
    )
VALUES ($1, $2, $3, $4)
RETURNING *;
-- name: QueryGroupMessagesAsc :many
SELECT *
FROM group_messages
WHERE group_id = @group_id
ORDER BY id ASC
LIMIT @numrows;
-- name: QueryGroupMessagesDesc :many
SELECT *
FROM group_messages
WHERE group_id = @group_id
ORDER BY id DESC
LIMIT @numrows;
-- name: QueryGroupMessagesWithCursorAsc :many
SELECT *
FROM group_messages
WHERE group_id = $1
    AND id > $2
ORDER BY id ASC
LIMIT $3;
-- name: QueryGroupMessagesWithCursorDesc :many
SELECT *
FROM group_messages
WHERE group_id = $1
    AND id < $2
ORDER BY id DESC
LIMIT $3;