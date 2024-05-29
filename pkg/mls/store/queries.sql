-- name: LockInboxLog :exec
SELECT
	pg_advisory_xact_lock(hashtext(@inbox_id));

-- name: GetAllInboxLogs :many
SELECT
	*
FROM
	inbox_log
WHERE
	inbox_id = $1
ORDER BY
	sequence_id ASC;

-- name: GetInboxLogFiltered :many
SELECT
	a.*
FROM
	inbox_log AS a
	JOIN (
		SELECT
			*
		FROM
			json_populate_recordset(NULL::inbox_filter, @filters) AS b(inbox_id,
				sequence_id)) AS b ON decode(b.inbox_id, 'hex') = a.inbox_id::BYTEA
		AND a.sequence_id > b.sequence_id
	ORDER BY
		a.sequence_id ASC;

-- name: GetAddressLogs :many
SELECT
	a.address,
	a.inbox_id,
	a.association_sequence_id
FROM
	address_log a
	INNER JOIN (
		SELECT
			address,
			MAX(association_sequence_id) AS max_association_sequence_id
		FROM
			address_log
		WHERE
			address = ANY (@addresses::TEXT[])
			AND revocation_sequence_id IS NULL
		GROUP BY
			address) b ON a.address = b.address
	AND a.association_sequence_id = b.max_association_sequence_id;

-- name: InsertAddressLog :one
INSERT INTO address_log(address, inbox_id, association_sequence_id, revocation_sequence_id)
	VALUES ($1, $2, $3, $4)
RETURNING
	*;

-- name: InsertInboxLog :one
INSERT INTO inbox_log(inbox_id, server_timestamp_ns, identity_update_proto)
	VALUES ($1, $2, $3)
RETURNING
	sequence_id;

-- name: RevokeAddressFromLog :exec
UPDATE
	address_log
SET
	revocation_sequence_id = $1
WHERE (address, inbox_id, association_sequence_id) =(
	SELECT
		address,
		inbox_id,
		MAX(association_sequence_id)
	FROM
		address_log AS a
	WHERE
		a.address = $2
		AND a.inbox_id = $3
	GROUP BY
		address,
		inbox_id);

-- name: CreateInstallation :exec
INSERT INTO installations(id, created_at, updated_at, inbox_id, key_package, expiration)
	VALUES (@id, @created_at, @updated_at, @inbox_id, @key_package, @expiration);

-- name: GetInstallation :one
SELECT
	*
FROM
	installations
WHERE
	id = $1;

-- name: UpdateKeyPackage :execrows
UPDATE
	installations
SET
	key_package = @key_package,
	updated_at = @updated_at,
	expiration = @expiration
WHERE
	id = @id;

-- name: FetchKeyPackages :many
SELECT
	id,
	key_package
FROM
	installations
WHERE
	id = ANY (@installation_ids::BYTEA[]);

-- name: InsertGroupMessage :one
INSERT INTO group_messages(group_id, data, group_id_data_hash)
	VALUES ($1, $2, $3)
RETURNING
	*;

-- name: InsertWelcomeMessage :one
INSERT INTO welcome_messages(installation_key, data, installation_key_data_hash, hpke_public_key)
	VALUES ($1, $2, $3, $4)
RETURNING
	*;

-- name: GetAllGroupMessages :many
SELECT
	*
FROM
	group_messages
ORDER BY
	id ASC;

-- name: QueryGroupMessages :many
SELECT
	*
FROM
	group_messages
WHERE
	group_id = @group_id
ORDER BY
	CASE WHEN @sort_desc::BOOL THEN
		id
	END DESC,
	CASE WHEN @sort_desc::BOOL = FALSE THEN
		id
	END ASC
LIMIT @numrows;

-- name: QueryGroupMessagesWithCursorAsc :many
SELECT
	*
FROM
	group_messages
WHERE
	group_id = @group_id
	AND id > @cursor
ORDER BY
	id ASC
LIMIT @numrows;

-- name: QueryGroupMessagesWithCursorDesc :many
SELECT
	*
FROM
	group_messages
WHERE
	group_id = @group_id
	AND id < @cursor
ORDER BY
	id DESC
LIMIT @numrows;

-- name: GetAllWelcomeMessages :many
SELECT
	*
FROM
	welcome_messages
ORDER BY
	id ASC;

-- name: QueryWelcomeMessages :many
SELECT
	*
FROM
	welcome_messages
WHERE
	installation_key = @installation_key
ORDER BY
	CASE WHEN @sort_desc::BOOL THEN
		id
	END DESC,
	CASE WHEN @sort_desc::BOOL = FALSE THEN
		id
	END ASC
LIMIT @numrows;

-- name: QueryWelcomeMessagesWithCursorAsc :many
SELECT
	*
FROM
	welcome_messages
WHERE
	installation_key = @installation_key
	AND id > @cursor
ORDER BY
	id ASC
LIMIT @numrows;

-- name: QueryWelcomeMessagesWithCursorDesc :many
SELECT
	*
FROM
	welcome_messages
WHERE
	installation_key = @installation_key
	AND id < @cursor
ORDER BY
	id DESC
LIMIT @numrows;

