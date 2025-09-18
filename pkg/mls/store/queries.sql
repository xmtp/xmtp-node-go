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
	a.address,
	encode(a.inbox_id, 'hex') AS inbox_id,
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
	VALUES (@address, decode(@inbox_id, 'hex'), @association_sequence_id, @revocation_sequence_id)
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
WHERE (address, inbox_id, association_sequence_id) =(
	SELECT
		address,
		inbox_id,
		MAX(association_sequence_id)
	FROM
		address_log AS a
	WHERE
		a.address = @address
		AND a.inbox_id = decode(@inbox_id, 'hex')
	GROUP BY
		address,
		inbox_id);

-- name: CreateOrUpdateInstallation :exec
INSERT INTO installations(id, created_at, updated_at, key_package, is_appended)
	VALUES (@id, @created_at, @updated_at, @key_package, @is_appended)
ON CONFLICT (id)
	DO UPDATE SET
		key_package = @key_package,
		updated_at = @updated_at;

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
	insert_group_message_v3(@group_id, @data, @group_id_data_hash, @sender_hmac, @should_push, @is_commit);

-- name: InsertWelcomeMessage :one
SELECT
	*
FROM
	insert_welcome_message_v4(@installation_key, @data, @installation_key_data_hash, @hpke_public_key, @wrapper_algorithm, @welcome_metadata);

-- name: InsertWelcomePointerMessage :one
SELECT
	id, created_at, installation_key, data, hpke_public_key, installation_key_data_hash, wrapper_algorithm, welcome_metadata, message_type
FROM
	insert_welcome_pointer_message_v1(@installation_key, @welcome_pointer_data, @installation_key_data_hash, @hpke_public_key, @wrapper_algorithm);

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

-- name: QueryCommitLogV2 :many
SELECT
	*
FROM
	commit_log_v2
WHERE
	group_id = @group_id
	AND id > @cursor
ORDER BY
	id ASC
LIMIT @numrows;

-- name: TouchInbox :exec
INSERT INTO inboxes(id)
	VALUES (decode(@inbox_id, 'hex'))
ON CONFLICT (id)
	DO UPDATE SET
		updated_at = NOW();

-- name: GetOldWelcomeMessages :one
SELECT
	COUNT(*)::BIGINT AS old_message_count
FROM
	welcome_messages
WHERE
	created_at < NOW() - make_interval(days := @age_days);

-- name: DeleteOldWelcomeMessagesBatch :many
WITH to_delete AS (
	SELECT
		id
	FROM
		welcome_messages
	WHERE
		created_at < NOW() - make_interval(days := @age_days)
	ORDER BY
		id
	LIMIT @batch_size
	FOR UPDATE
		SKIP LOCKED)
DELETE FROM welcome_messages wm USING to_delete td
WHERE wm.id = td.id
RETURNING
	wm.id, wm.created_at;

-- name: InsertCommitLogV2 :one
SELECT
	*
FROM
	insert_commit_log_v2(@group_id, @serialized_entry, @serialized_signature);

-- name: CountDeletableGroupMessages :one
SELECT
	COUNT(*)
FROM
	group_messages
WHERE
	is_commit = FALSE
	AND created_at < NOW() - make_interval(days := @age_days);

-- name: DeleteOldGroupMessagesBatch :many
WITH to_delete AS (
	SELECT
		id
	FROM
		group_messages
	WHERE
		is_commit = FALSE
		AND created_at < NOW() - make_interval(days := @age_days)
	ORDER BY
		id
	LIMIT @batch_size
	FOR UPDATE
		SKIP LOCKED)
DELETE FROM group_messages gm USING to_delete td
WHERE gm.id = td.id
RETURNING
	gm.id, gm.created_at;

-- name: GetOldInstallations :one
SELECT
	COUNT(*)::BIGINT AS old_message_count
FROM
	installations
WHERE
	created_at < NOW() - make_interval(days := @age_days);

-- name: DeleteOldInstallationsBatch :many
WITH to_delete AS (
	SELECT
		id
	FROM
		installations
	WHERE
		created_at <(EXTRACT(EPOCH FROM NOW() -((@age_days)::INT || ' days')::INTERVAL) * 1e9)::BIGINT
	ORDER BY
		id
	LIMIT @batch_size
	FOR UPDATE
		SKIP LOCKED)
DELETE FROM installations i USING to_delete td
WHERE i.id = td.id
RETURNING
	i.id, i.created_at;

-- name: SelectEnvelopesForIsCommitBackfill :many
SELECT
	id,
	data
FROM
	group_messages
WHERE
	is_commit IS NULL
ORDER BY
	id ASC
FOR UPDATE
	SKIP LOCKED
LIMIT 100;

-- name: UpdateIsCommitStatus :exec
UPDATE
	group_messages
SET
	is_commit = @is_commit
WHERE
	id = @id;

-- name: GetAllGroupMessagesWithCursor :many
SELECT
	*
FROM
	group_messages
WHERE
	id > @cursor
ORDER BY
	id ASC
LIMIT @numrows;

-- name: GetAllWelcomeMessagesWithCursor :many
SELECT
	*
FROM
	welcome_messages
WHERE
	id > @cursor
ORDER BY
	id ASC
LIMIT @numrows;

-- name: GetLatestGroupMessageID :one
SELECT
	COALESCE((
		SELECT
			max(id)
		FROM group_messages), 0)::BIGINT
LIMIT 1;

-- name: GetLatestWelcomeMessageID :one
SELECT
	COALESCE((
		SELECT
			max(id)
		FROM welcome_messages), 0)::BIGINT
LIMIT 1;

-- name: SelectInstallationsToBackfill :many
SELECT
	id,
	key_package,
	created_at
FROM
	installations
WHERE
	is_appended IS NULL
ORDER BY
	id ASC
FOR UPDATE
	SKIP LOCKED
LIMIT 100;

-- name: UpdateIsAppendedStatus :exec
UPDATE
	installations
SET
	is_appended = @is_appended
WHERE
	id = @id;

-- name: InsertKeyPackage :exec
INSERT INTO key_packages(installation_id, key_package, created_at)
	VALUES (@installation_id, @key_package, @created_at)
ON CONFLICT (installation_id, key_package)
	DO NOTHING;

-- name: GetOldKeyPackages :one
SELECT
	COUNT(*)::BIGINT AS old_message_count
FROM
	key_packages
WHERE
	created_at < NOW() - make_interval(days := @age_days);

-- name: DeleteOldKeyPackagesBatch :many
WITH to_delete AS (
	SELECT
		installation_id
	FROM
		key_packages
	WHERE
		created_at <(EXTRACT(EPOCH FROM NOW() -((@age_days)::INT || ' days')::INTERVAL) * 1e9)::BIGINT
	ORDER BY
		installation_id
	LIMIT @batch_size
	FOR UPDATE
		SKIP LOCKED)
DELETE FROM key_packages kp USING to_delete td
WHERE kp.installation_id = td.installation_id
RETURNING
	kp.installation_id, kp.created_at;

-- name: GetNewestGroupMessage :many
SELECT DISTINCT ON (group_id)
	id,
	group_id,
	data,
	created_at,
	should_push,
	sender_hmac,
	is_commit
FROM
	group_messages
WHERE
	group_id = ANY (@group_ids::BYTEA[])
ORDER BY
	group_id,
	id DESC;

-- name: GetNewestGroupMessageMetadata :many
SELECT DISTINCT ON (group_id)
	id,
	group_id,
	created_at,
	is_commit
FROM
	group_messages
WHERE
	group_id = ANY (@group_ids::BYTEA[])
ORDER BY
	group_id,
	id DESC;
