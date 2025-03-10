-- name: LockInboxLog :exec
SELECT
	pg_advisory_xact_lock(hashtext(@inbox_id));

-- name: GetAllInboxLogs :many
SELECT
	sequence_id,
	encode(inbox_id, 'hex') AS inbox_id,
	identity_update_proto
FROM
	inbox_log
WHERE
	inbox_id = decode(@inbox_id, 'hex')
ORDER BY
	sequence_id ASC;

-- name: GetInboxLogFiltered :many
SELECT
	a.sequence_id,
	encode(a.inbox_id, 'hex') AS inbox_id,
	a.identity_update_proto,
	a.server_timestamp_ns
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
	a.identifier,
	a.identifier_kind,
	encode(a.inbox_id, 'hex') AS inbox_id,
	a.association_sequence_id
FROM
	address_log a
	INNER JOIN (
		SELECT
			identifier,
			identifier_kind,
			MAX(association_sequence_id) AS max_association_sequence_id
		FROM
			address_log
		WHERE
		    (identifier, identifier_kind) IN (SELECT unnest(@identifiers::TEXT[]), unnest(@identifier_kinds::INT[]))
			AND revocation_sequence_id IS NULL
		GROUP BY
			identifier) b ON a.identifier = b.identifier
	AND a.association_sequence_id = b.max_association_sequence_id;

-- name: InsertAddressLog :one
INSERT INTO address_log(identifier, identifier_kind, inbox_id, association_sequence_id, revocation_sequence_id)
	VALUES (@identifier, @identifier_kind, decode(@inbox_id, 'hex'), @association_sequence_id, @revocation_sequence_id)
RETURNING
	*;

-- name: InsertInboxLog :one
SELECT
	sequence_id
FROM
	insert_inbox_log(decode(@inbox_id, 'hex'), @server_timestamp_ns, @identity_update_proto);

-- name: RevokeAddressFromLog :exec
UPDATE
	address_log
SET
	revocation_sequence_id = @revocation_sequence_id
WHERE (identifier, identifier_kind, inbox_id, association_sequence_id) =(
	SELECT
		identifier,
		identifier_kind,
		inbox_id,
		MAX(association_sequence_id)
	FROM
		address_log AS a
	WHERE
		a.identifier = @identifier
		AND a.identifier_kind = @identifier_kind,
		AND a.inbox_id = decode(@inbox_id, 'hex')
	GROUP BY
		identifier,
		identifier_kind,
		inbox_id);

-- name: CreateOrUpdateInstallation :exec
INSERT INTO installations(id, created_at, updated_at, key_package)
	VALUES (@id, @created_at, @updated_at, @key_package)
ON CONFLICT (id)
	DO UPDATE SET
		key_package = @key_package, updated_at = @updated_at;

-- name: GetInstallation :one
SELECT
	id,
	created_at,
	updated_at,
	key_package
FROM
	installations
WHERE
	id = $1;

-- name: FetchKeyPackages :many
SELECT
	id,
	key_package
FROM
	installations
WHERE
	id = ANY (@installation_ids::BYTEA[]);

-- name: InsertGroupMessage :one
SELECT
	*
FROM
	insert_group_message(@group_id, @data, @group_id_data_hash);

-- name: InsertWelcomeMessage :one
SELECT
	*
FROM
	insert_welcome_message(@installation_key, @data, @installation_key_data_hash, @hpke_public_key);

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
