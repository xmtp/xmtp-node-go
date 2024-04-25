// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: queries.sql

package queries

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/lib/pq"
)

const createInstallation = `-- name: CreateInstallation :exec
INSERT INTO installations (id, wallet_address, created_at, updated_at, credential_identity, key_package, expiration)
VALUES ($1, $2, $3, $3, $4, $5, $6)
`

type CreateInstallationParams struct {
	ID                 []byte
	WalletAddress      string
	CreatedAt          int64
	CredentialIdentity []byte
	KeyPackage         []byte
	Expiration         int64
}

func (q *Queries) CreateInstallation(ctx context.Context, arg CreateInstallationParams) error {
	_, err := q.db.ExecContext(ctx, createInstallation,
		arg.ID,
		arg.WalletAddress,
		arg.CreatedAt,
		arg.CredentialIdentity,
		arg.KeyPackage,
		arg.Expiration,
	)
	return err
}

const fetchKeyPackages = `-- name: FetchKeyPackages :many
SELECT id, key_package FROM installations
WHERE id = ANY ($1::bytea[])
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
        address = ANY ($1::text[])
    AND
        revocation_sequence_id IS NULL
    GROUP BY 
        address
) b ON a.address = b.address AND a.association_sequence_id = b.max_association_sequence_id
`

type GetAddressLogsRow struct {
	Address               string
	InboxID               string
	AssociationSequenceID sql.NullInt64
}

func (q *Queries) GetAddressLogs(ctx context.Context, addresses []string) ([]GetAddressLogsRow, error) {
	rows, err := q.db.QueryContext(ctx, getAddressLogs, pq.Array(addresses))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetAddressLogsRow
	for rows.Next() {
		var i GetAddressLogsRow
		if err := rows.Scan(&i.Address, &i.InboxID, &i.AssociationSequenceID); err != nil {
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
SELECT sequence_id, inbox_id, server_timestamp_ns, identity_update_proto FROM inbox_log
WHERE inbox_id = $1
ORDER BY sequence_id ASC FOR UPDATE
`

func (q *Queries) GetAllInboxLogs(ctx context.Context, inboxID string) ([]InboxLog, error) {
	rows, err := q.db.QueryContext(ctx, getAllInboxLogs, inboxID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []InboxLog
	for rows.Next() {
		var i InboxLog
		if err := rows.Scan(
			&i.SequenceID,
			&i.InboxID,
			&i.ServerTimestampNs,
			&i.IdentityUpdateProto,
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

const getIdentityUpdates = `-- name: GetIdentityUpdates :many
SELECT id, wallet_address, created_at, updated_at, credential_identity, revoked_at, key_package, expiration FROM installations
WHERE wallet_address = ANY ($1::text[])
AND (created_at > $2 OR revoked_at > $2)
ORDER BY created_at ASC
`

type GetIdentityUpdatesParams struct {
	WalletAddresses []string
	StartTime       int64
}

func (q *Queries) GetIdentityUpdates(ctx context.Context, arg GetIdentityUpdatesParams) ([]Installation, error) {
	rows, err := q.db.QueryContext(ctx, getIdentityUpdates, pq.Array(arg.WalletAddresses), arg.StartTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Installation
	for rows.Next() {
		var i Installation
		if err := rows.Scan(
			&i.ID,
			&i.WalletAddress,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.CredentialIdentity,
			&i.RevokedAt,
			&i.KeyPackage,
			&i.Expiration,
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
SELECT a.sequence_id, a.inbox_id, a.server_timestamp_ns, a.identity_update_proto FROM inbox_log AS a
JOIN (
    SELECT inbox_id, sequence_id FROM json_populate_recordset(null::inbox_filter, $1) as b(inbox_id, sequence_id)
) as b on b.inbox_id = a.inbox_id AND a.sequence_id > b.sequence_id
ORDER BY a.sequence_id ASC
`

func (q *Queries) GetInboxLogFiltered(ctx context.Context, filters json.RawMessage) ([]InboxLog, error) {
	rows, err := q.db.QueryContext(ctx, getInboxLogFiltered, filters)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []InboxLog
	for rows.Next() {
		var i InboxLog
		if err := rows.Scan(
			&i.SequenceID,
			&i.InboxID,
			&i.ServerTimestampNs,
			&i.IdentityUpdateProto,
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

const insertAddressLog = `-- name: InsertAddressLog :one
INSERT INTO address_log (address, inbox_id, association_sequence_id, revocation_sequence_id)
VALUES ($1, $2, $3, $4)
RETURNING address, inbox_id, association_sequence_id, revocation_sequence_id
`

type InsertAddressLogParams struct {
	Address               string
	InboxID               string
	AssociationSequenceID sql.NullInt64
	RevocationSequenceID  sql.NullInt64
}

func (q *Queries) InsertAddressLog(ctx context.Context, arg InsertAddressLogParams) (AddressLog, error) {
	row := q.db.QueryRowContext(ctx, insertAddressLog,
		arg.Address,
		arg.InboxID,
		arg.AssociationSequenceID,
		arg.RevocationSequenceID,
	)
	var i AddressLog
	err := row.Scan(
		&i.Address,
		&i.InboxID,
		&i.AssociationSequenceID,
		&i.RevocationSequenceID,
	)
	return i, err
}

const insertGroupMessage = `-- name: InsertGroupMessage :one
INSERT INTO group_messages (group_id, data, group_id_data_hash)
VALUES ($1, $2, $3)
RETURNING id, created_at, group_id, data, group_id_data_hash
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
INSERT INTO inbox_log (inbox_id, server_timestamp_ns, identity_update_proto)
VALUES ($1, $2, $3)
RETURNING sequence_id
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
INSERT INTO welcome_messages (installation_key, data, installation_key_data_hash, hpke_public_key)
VALUES ($1, $2, $3, $4)
RETURNING id, created_at, installation_key, data, installation_key_data_hash, hpke_public_key
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
		&i.InstallationKeyDataHash,
		&i.HpkePublicKey,
	)
	return i, err
}

const queryGroupMessagesAsc = `-- name: QueryGroupMessagesAsc :many
SELECT id, created_at, group_id, data, group_id_data_hash FROM group_messages
WHERE group_id = $1
ORDER BY id ASC
LIMIT $2
`

type QueryGroupMessagesAscParams struct {
	GroupID []byte
	Numrows int32
}

func (q *Queries) QueryGroupMessagesAsc(ctx context.Context, arg QueryGroupMessagesAscParams) ([]GroupMessage, error) {
	rows, err := q.db.QueryContext(ctx, queryGroupMessagesAsc, arg.GroupID, arg.Numrows)
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

const queryGroupMessagesDesc = `-- name: QueryGroupMessagesDesc :many
SELECT id, created_at, group_id, data, group_id_data_hash FROM group_messages
WHERE group_id = $1
ORDER BY id DESC
LIMIT $2
`

type QueryGroupMessagesDescParams struct {
	GroupID []byte
	Numrows int32
}

func (q *Queries) QueryGroupMessagesDesc(ctx context.Context, arg QueryGroupMessagesDescParams) ([]GroupMessage, error) {
	rows, err := q.db.QueryContext(ctx, queryGroupMessagesDesc, arg.GroupID, arg.Numrows)
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
SELECT id, created_at, group_id, data, group_id_data_hash FROM group_messages
WHERE group_id = $1
AND id > $2
ORDER BY id ASC
LIMIT $3
`

type QueryGroupMessagesWithCursorAscParams struct {
	GroupID []byte
	ID      int64
	Limit   int32
}

func (q *Queries) QueryGroupMessagesWithCursorAsc(ctx context.Context, arg QueryGroupMessagesWithCursorAscParams) ([]GroupMessage, error) {
	rows, err := q.db.QueryContext(ctx, queryGroupMessagesWithCursorAsc, arg.GroupID, arg.ID, arg.Limit)
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
SELECT id, created_at, group_id, data, group_id_data_hash FROM group_messages
WHERE group_id = $1
AND id < $2
ORDER BY id DESC
LIMIT $3
`

type QueryGroupMessagesWithCursorDescParams struct {
	GroupID []byte
	ID      int64
	Limit   int32
}

func (q *Queries) QueryGroupMessagesWithCursorDesc(ctx context.Context, arg QueryGroupMessagesWithCursorDescParams) ([]GroupMessage, error) {
	rows, err := q.db.QueryContext(ctx, queryGroupMessagesWithCursorDesc, arg.GroupID, arg.ID, arg.Limit)
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

const revokeAddressFromLog = `-- name: RevokeAddressFromLog :exec
UPDATE address_log
SET revocation_sequence_id = $1
WHERE (address, inbox_id, association_sequence_id) = (
    SELECT address, inbox_id, MAX(association_sequence_id)
    FROM address_log AS a
    WHERE a.address = $2 AND a.inbox_id = $3
    GROUP BY address, inbox_id
)
`

type RevokeAddressFromLogParams struct {
	RevocationSequenceID sql.NullInt64
	Address              string
	InboxID              string
}

func (q *Queries) RevokeAddressFromLog(ctx context.Context, arg RevokeAddressFromLogParams) error {
	_, err := q.db.ExecContext(ctx, revokeAddressFromLog, arg.RevocationSequenceID, arg.Address, arg.InboxID)
	return err
}

const revokeInstallation = `-- name: RevokeInstallation :exec
UPDATE installations
SET revoked_at = $1
WHERE id = $2
AND revoked_at IS NULL
`

type RevokeInstallationParams struct {
	RevokedAt      sql.NullInt64
	InstallationID []byte
}

func (q *Queries) RevokeInstallation(ctx context.Context, arg RevokeInstallationParams) error {
	_, err := q.db.ExecContext(ctx, revokeInstallation, arg.RevokedAt, arg.InstallationID)
	return err
}

const updateKeyPackage = `-- name: UpdateKeyPackage :execrows
UPDATE installations
SET key_package = $1, updated_at = $2, expiration = $3
WHERE id = $4
`

type UpdateKeyPackageParams struct {
	KeyPackage []byte
	UpdatedAt  int64
	Expiration int64
	ID         []byte
}

func (q *Queries) UpdateKeyPackage(ctx context.Context, arg UpdateKeyPackageParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, updateKeyPackage,
		arg.KeyPackage,
		arg.UpdatedAt,
		arg.Expiration,
		arg.ID,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
