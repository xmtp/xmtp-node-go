// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: queries.sql

package queries

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/lib/pq"
)

const createOrUpdateInstallation = `-- name: CreateOrUpdateInstallation :exec
INSERT INTO installations(id, created_at, updated_at, key_package)
	VALUES ($1, $2, $3, $4)
ON CONFLICT (id)
	DO UPDATE SET
		key_package = $4, updated_at = $3
`

type CreateOrUpdateInstallationParams struct {
	ID         []byte
	CreatedAt  int64
	UpdatedAt  int64
	KeyPackage []byte
}

func (q *Queries) CreateOrUpdateInstallation(ctx context.Context, arg CreateOrUpdateInstallationParams) error {
	_, err := q.db.ExecContext(ctx, createOrUpdateInstallation,
		arg.ID,
		arg.CreatedAt,
		arg.UpdatedAt,
		arg.KeyPackage,
	)
	return err
}

const fetchKeyPackages = `-- name: FetchKeyPackages :many
SELECT
	id,
	key_package
FROM
	installations
WHERE
	id = ANY ($1::BYTEA[])
`

type FetchKeyPackagesRow struct {
	ID         []byte
	KeyPackage []byte
}

func (q *Queries) FetchKeyPackages(ctx context.Context, installationIds [][]byte) ([]FetchKeyPackagesRow, error) {
	rows, err := q.db.QueryContext(ctx, fetchKeyPackages, pq.Array(installationIds))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []FetchKeyPackagesRow
	for rows.Next() {
		var i FetchKeyPackagesRow
		if err := rows.Scan(&i.ID, &i.KeyPackage); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getAddressLogs = `-- name: GetAddressLogs :many
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
		    (identifier, identifier_kind) IN (SELECT unnest($1::TEXT[]), unnest($2::INT[]))
			AND revocation_sequence_id IS NULL
		GROUP BY
			identifier, identifier_kind) b ON a.identifier = b.identifier
	AND a.association_sequence_id = b.max_association_sequence_id
`

type GetAddressLogsParams struct {
	Identifiers     []string
	IdentifierKinds []int32
}

type GetAddressLogsRow struct {
	Identifier            string
	IdentifierKind        int32
	InboxID               string
	AssociationSequenceID sql.NullInt64
}

func (q *Queries) GetAddressLogs(ctx context.Context, arg GetAddressLogsParams) ([]GetAddressLogsRow, error) {
	rows, err := q.db.QueryContext(ctx, getAddressLogs, pq.Array(arg.Identifiers), pq.Array(arg.IdentifierKinds))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetAddressLogsRow
	for rows.Next() {
		var i GetAddressLogsRow
		if err := rows.Scan(
			&i.Identifier,
			&i.IdentifierKind,
			&i.InboxID,
			&i.AssociationSequenceID,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getAllGroupMessages = `-- name: GetAllGroupMessages :many
SELECT
	id, created_at, group_id, data, group_id_data_hash
FROM
	group_messages
ORDER BY
	id ASC
`

func (q *Queries) GetAllGroupMessages(ctx context.Context) ([]GroupMessage, error) {
	rows, err := q.db.QueryContext(ctx, getAllGroupMessages)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GroupMessage
	for rows.Next() {
		var i GroupMessage
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.GroupID,
			&i.Data,
			&i.GroupIDDataHash,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getAllInboxLogs = `-- name: GetAllInboxLogs :many
SELECT
	sequence_id,
	encode(inbox_id, 'hex') AS inbox_id,
	identity_update_proto
FROM
	inbox_log
WHERE
	inbox_id = decode($1, 'hex')
ORDER BY
	sequence_id ASC
`

type GetAllInboxLogsRow struct {
	SequenceID          int64
	InboxID             string
	IdentityUpdateProto []byte
}

func (q *Queries) GetAllInboxLogs(ctx context.Context, inboxID string) ([]GetAllInboxLogsRow, error) {
	rows, err := q.db.QueryContext(ctx, getAllInboxLogs, inboxID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetAllInboxLogsRow
	for rows.Next() {
		var i GetAllInboxLogsRow
		if err := rows.Scan(&i.SequenceID, &i.InboxID, &i.IdentityUpdateProto); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getAllWelcomeMessages = `-- name: GetAllWelcomeMessages :many
SELECT
	id, created_at, installation_key, data, hpke_public_key, installation_key_data_hash
FROM
	welcome_messages
ORDER BY
	id ASC
`

func (q *Queries) GetAllWelcomeMessages(ctx context.Context) ([]WelcomeMessage, error) {
	rows, err := q.db.QueryContext(ctx, getAllWelcomeMessages)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []WelcomeMessage
	for rows.Next() {
		var i WelcomeMessage
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.InstallationKey,
			&i.Data,
			&i.HpkePublicKey,
			&i.InstallationKeyDataHash,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getInboxLogFiltered = `-- name: GetInboxLogFiltered :many
SELECT
	a.sequence_id,
	encode(a.inbox_id, 'hex') AS inbox_id,
	a.identity_update_proto,
	a.server_timestamp_ns
FROM
	inbox_log AS a
	JOIN (
		SELECT
			inbox_id, sequence_id
		FROM
			json_populate_recordset(NULL::inbox_filter, $1) AS b(inbox_id,
				sequence_id)) AS b ON decode(b.inbox_id, 'hex') = a.inbox_id::BYTEA
		AND a.sequence_id > b.sequence_id
	ORDER BY
		a.sequence_id ASC
`

type GetInboxLogFilteredRow struct {
	SequenceID          int64
	InboxID             string
	IdentityUpdateProto []byte
	ServerTimestampNs   int64
}

func (q *Queries) GetInboxLogFiltered(ctx context.Context, filters json.RawMessage) ([]GetInboxLogFilteredRow, error) {
	rows, err := q.db.QueryContext(ctx, getInboxLogFiltered, filters)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetInboxLogFilteredRow
	for rows.Next() {
		var i GetInboxLogFilteredRow
		if err := rows.Scan(
			&i.SequenceID,
			&i.InboxID,
			&i.IdentityUpdateProto,
			&i.ServerTimestampNs,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getInstallation = `-- name: GetInstallation :one
SELECT
	id,
	created_at,
	updated_at,
	key_package
FROM
	installations
WHERE
	id = $1
`

func (q *Queries) GetInstallation(ctx context.Context, id []byte) (Installation, error) {
	row := q.db.QueryRowContext(ctx, getInstallation, id)
	var i Installation
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.KeyPackage,
	)
	return i, err
}

const insertAddressLog = `-- name: InsertAddressLog :one
INSERT INTO address_log(identifier, identifier_kind, inbox_id, association_sequence_id, revocation_sequence_id)
	VALUES ($1, $2, decode($3, 'hex'), $4, $5)
RETURNING
	identifier, inbox_id, association_sequence_id, revocation_sequence_id, identifier_kind
`

type InsertAddressLogParams struct {
	Identifier            string
	IdentifierKind        int32
	InboxID               string
	AssociationSequenceID sql.NullInt64
	RevocationSequenceID  sql.NullInt64
}

func (q *Queries) InsertAddressLog(ctx context.Context, arg InsertAddressLogParams) (AddressLog, error) {
	row := q.db.QueryRowContext(ctx, insertAddressLog,
		arg.Identifier,
		arg.IdentifierKind,
		arg.InboxID,
		arg.AssociationSequenceID,
		arg.RevocationSequenceID,
	)
	var i AddressLog
	err := row.Scan(
		&i.Identifier,
		&i.InboxID,
		&i.AssociationSequenceID,
		&i.RevocationSequenceID,
		&i.IdentifierKind,
	)
	return i, err
}

const insertGroupMessage = `-- name: InsertGroupMessage :one
SELECT
	id, created_at, group_id, data, group_id_data_hash
FROM
	insert_group_message($1, $2, $3)
`

type InsertGroupMessageParams struct {
	GroupID         []byte
	Data            []byte
	GroupIDDataHash []byte
}

func (q *Queries) InsertGroupMessage(ctx context.Context, arg InsertGroupMessageParams) (GroupMessage, error) {
	row := q.db.QueryRowContext(ctx, insertGroupMessage, arg.GroupID, arg.Data, arg.GroupIDDataHash)
	var i GroupMessage
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.GroupID,
		&i.Data,
		&i.GroupIDDataHash,
	)
	return i, err
}

const insertInboxLog = `-- name: InsertInboxLog :one
SELECT
	sequence_id
FROM
	insert_inbox_log(decode($1, 'hex'), $2, $3)
`

type InsertInboxLogParams struct {
	InboxID             string
	ServerTimestampNs   int64
	IdentityUpdateProto []byte
}

func (q *Queries) InsertInboxLog(ctx context.Context, arg InsertInboxLogParams) (int64, error) {
	row := q.db.QueryRowContext(ctx, insertInboxLog, arg.InboxID, arg.ServerTimestampNs, arg.IdentityUpdateProto)
	var sequence_id int64
	err := row.Scan(&sequence_id)
	return sequence_id, err
}

const insertWelcomeMessage = `-- name: InsertWelcomeMessage :one
SELECT
	id, created_at, installation_key, data, hpke_public_key, installation_key_data_hash
FROM
	insert_welcome_message($1, $2, $3, $4)
`

type InsertWelcomeMessageParams struct {
	InstallationKey         []byte
	Data                    []byte
	InstallationKeyDataHash []byte
	HpkePublicKey           []byte
}

func (q *Queries) InsertWelcomeMessage(ctx context.Context, arg InsertWelcomeMessageParams) (WelcomeMessage, error) {
	row := q.db.QueryRowContext(ctx, insertWelcomeMessage,
		arg.InstallationKey,
		arg.Data,
		arg.InstallationKeyDataHash,
		arg.HpkePublicKey,
	)
	var i WelcomeMessage
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.InstallationKey,
		&i.Data,
		&i.HpkePublicKey,
		&i.InstallationKeyDataHash,
	)
	return i, err
}

const lockInboxLog = `-- name: LockInboxLog :exec
SELECT
	pg_advisory_xact_lock(hashtext($1))
`

func (q *Queries) LockInboxLog(ctx context.Context, inboxID string) error {
	_, err := q.db.ExecContext(ctx, lockInboxLog, inboxID)
	return err
}

const queryGroupMessages = `-- name: QueryGroupMessages :many
SELECT
	id, created_at, group_id, data, group_id_data_hash
FROM
	group_messages
WHERE
	group_id = $1
ORDER BY
	CASE WHEN $2::BOOL THEN
		id
	END DESC,
	CASE WHEN $2::BOOL = FALSE THEN
		id
	END ASC
LIMIT $3
`

type QueryGroupMessagesParams struct {
	GroupID  []byte
	SortDesc bool
	Numrows  int32
}

func (q *Queries) QueryGroupMessages(ctx context.Context, arg QueryGroupMessagesParams) ([]GroupMessage, error) {
	rows, err := q.db.QueryContext(ctx, queryGroupMessages, arg.GroupID, arg.SortDesc, arg.Numrows)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GroupMessage
	for rows.Next() {
		var i GroupMessage
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.GroupID,
			&i.Data,
			&i.GroupIDDataHash,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const queryGroupMessagesWithCursorAsc = `-- name: QueryGroupMessagesWithCursorAsc :many
SELECT
	id, created_at, group_id, data, group_id_data_hash
FROM
	group_messages
WHERE
	group_id = $1
	AND id > $2
ORDER BY
	id ASC
LIMIT $3
`

type QueryGroupMessagesWithCursorAscParams struct {
	GroupID []byte
	Cursor  int64
	Numrows int32
}

func (q *Queries) QueryGroupMessagesWithCursorAsc(ctx context.Context, arg QueryGroupMessagesWithCursorAscParams) ([]GroupMessage, error) {
	rows, err := q.db.QueryContext(ctx, queryGroupMessagesWithCursorAsc, arg.GroupID, arg.Cursor, arg.Numrows)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GroupMessage
	for rows.Next() {
		var i GroupMessage
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.GroupID,
			&i.Data,
			&i.GroupIDDataHash,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const queryGroupMessagesWithCursorDesc = `-- name: QueryGroupMessagesWithCursorDesc :many
SELECT
	id, created_at, group_id, data, group_id_data_hash
FROM
	group_messages
WHERE
	group_id = $1
	AND id < $2
ORDER BY
	id DESC
LIMIT $3
`

type QueryGroupMessagesWithCursorDescParams struct {
	GroupID []byte
	Cursor  int64
	Numrows int32
}

func (q *Queries) QueryGroupMessagesWithCursorDesc(ctx context.Context, arg QueryGroupMessagesWithCursorDescParams) ([]GroupMessage, error) {
	rows, err := q.db.QueryContext(ctx, queryGroupMessagesWithCursorDesc, arg.GroupID, arg.Cursor, arg.Numrows)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GroupMessage
	for rows.Next() {
		var i GroupMessage
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.GroupID,
			&i.Data,
			&i.GroupIDDataHash,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const queryWelcomeMessages = `-- name: QueryWelcomeMessages :many
SELECT
	id, created_at, installation_key, data, hpke_public_key, installation_key_data_hash
FROM
	welcome_messages
WHERE
	installation_key = $1
ORDER BY
	CASE WHEN $2::BOOL THEN
		id
	END DESC,
	CASE WHEN $2::BOOL = FALSE THEN
		id
	END ASC
LIMIT $3
`

type QueryWelcomeMessagesParams struct {
	InstallationKey []byte
	SortDesc        bool
	Numrows         int32
}

func (q *Queries) QueryWelcomeMessages(ctx context.Context, arg QueryWelcomeMessagesParams) ([]WelcomeMessage, error) {
	rows, err := q.db.QueryContext(ctx, queryWelcomeMessages, arg.InstallationKey, arg.SortDesc, arg.Numrows)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []WelcomeMessage
	for rows.Next() {
		var i WelcomeMessage
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.InstallationKey,
			&i.Data,
			&i.HpkePublicKey,
			&i.InstallationKeyDataHash,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const queryWelcomeMessagesWithCursorAsc = `-- name: QueryWelcomeMessagesWithCursorAsc :many
SELECT
	id, created_at, installation_key, data, hpke_public_key, installation_key_data_hash
FROM
	welcome_messages
WHERE
	installation_key = $1
	AND id > $2
ORDER BY
	id ASC
LIMIT $3
`

type QueryWelcomeMessagesWithCursorAscParams struct {
	InstallationKey []byte
	Cursor          int64
	Numrows         int32
}

func (q *Queries) QueryWelcomeMessagesWithCursorAsc(ctx context.Context, arg QueryWelcomeMessagesWithCursorAscParams) ([]WelcomeMessage, error) {
	rows, err := q.db.QueryContext(ctx, queryWelcomeMessagesWithCursorAsc, arg.InstallationKey, arg.Cursor, arg.Numrows)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []WelcomeMessage
	for rows.Next() {
		var i WelcomeMessage
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.InstallationKey,
			&i.Data,
			&i.HpkePublicKey,
			&i.InstallationKeyDataHash,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const queryWelcomeMessagesWithCursorDesc = `-- name: QueryWelcomeMessagesWithCursorDesc :many
SELECT
	id, created_at, installation_key, data, hpke_public_key, installation_key_data_hash
FROM
	welcome_messages
WHERE
	installation_key = $1
	AND id < $2
ORDER BY
	id DESC
LIMIT $3
`

type QueryWelcomeMessagesWithCursorDescParams struct {
	InstallationKey []byte
	Cursor          int64
	Numrows         int32
}

func (q *Queries) QueryWelcomeMessagesWithCursorDesc(ctx context.Context, arg QueryWelcomeMessagesWithCursorDescParams) ([]WelcomeMessage, error) {
	rows, err := q.db.QueryContext(ctx, queryWelcomeMessagesWithCursorDesc, arg.InstallationKey, arg.Cursor, arg.Numrows)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []WelcomeMessage
	for rows.Next() {
		var i WelcomeMessage
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.InstallationKey,
			&i.Data,
			&i.HpkePublicKey,
			&i.InstallationKeyDataHash,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const revokeAddressFromLog = `-- name: RevokeAddressFromLog :exec
UPDATE
	address_log
SET
	revocation_sequence_id = $1
WHERE (identifier, identifier_kind, inbox_id, association_sequence_id) =(
	SELECT
		identifier,
		identifier_kind,
		inbox_id,
		MAX(association_sequence_id)
	FROM
		address_log AS a
	WHERE
		a.identifier = $2
		AND a.identifier_kind = $3
		AND a.inbox_id = decode($4, 'hex')
	GROUP BY
		identifier,
		identifier_kind,
		inbox_id)
`

type RevokeAddressFromLogParams struct {
	RevocationSequenceID sql.NullInt64
	Identifier           string
	IdentifierKind       int32
	InboxID              string
}

func (q *Queries) RevokeAddressFromLog(ctx context.Context, arg RevokeAddressFromLogParams) error {
	_, err := q.db.ExecContext(ctx, revokeAddressFromLog,
		arg.RevocationSequenceID,
		arg.Identifier,
		arg.IdentifierKind,
		arg.InboxID,
	)
	return err
}
