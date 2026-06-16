package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
	"github.com/uptrace/bun"
	queries "github.com/xmtp/xmtp-node-go/pkg/mls/store/queries"
	identity "github.com/xmtp/xmtp-node-go/pkg/proto/identity/api/v1"
	"github.com/xmtp/xmtp-node-go/pkg/proto/identity/associations"
	mlsv1 "github.com/xmtp/xmtp-node-go/pkg/proto/mls/api/v1"
	"github.com/xmtp/xmtp-node-go/pkg/proto/mls/message_contents"
	"github.com/xmtp/xmtp-node-go/pkg/types"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type ReadStore struct {
	log     *zap.Logger
	db      *sql.DB
	queries *queries.Queries
}

func NewReadStore(log *zap.Logger, db *bun.DB) *ReadStore {
	return &ReadStore{
		log:     log.Named("read-mlsstore"),
		db:      db.DB,
		queries: queries.New(db.DB),
	}
}

func (s *ReadStore) Queries() *queries.Queries {
	return s.queries
}

func (s *ReadStore) GetInboxIds(
	ctx context.Context,
	req *identity.GetInboxIdsRequest,
) (*identity.GetInboxIdsResponse, error) {
	identifiers := []string{}
	for _, request := range req.Requests {
		identifiers = append(identifiers, request.GetIdentifier())
	}

	addressLogEntries, err := s.queries.GetAddressLogs(ctx, identifiers)
	if err != nil {
		return nil, err
	}

	out := make([]*identity.GetInboxIdsResponse_Response, len(identifiers))

	for index, identifier := range identifiers {
		resp := identity.GetInboxIdsResponse_Response{}
		resp.Identifier = identifier

		for _, logEntry := range addressLogEntries {
			if logEntry.Address == identifier {
				inboxId := logEntry.InboxID
				resp.InboxId = &inboxId
			}
		}
		out[index] = &resp
	}

	return &identity.GetInboxIdsResponse{
		Responses: out,
	}, nil
}

func (s *ReadStore) GetInboxLogs(
	ctx context.Context,
	batched_req *identity.GetIdentityUpdatesRequest,
) (*identity.GetIdentityUpdatesResponse, error) {
	reqs := batched_req.GetRequests()

	filters := make(queries.InboxLogFilterList, len(reqs))
	for i, req := range reqs {
		filters[i] = queries.InboxLogFilter{
			InboxId:    req.InboxId, // InboxLogFilters take inbox_id as text and decode it inside Postgres, since the filters are JSON
			SequenceId: int64(req.SequenceId),
		}
	}
	filterBytes, err := filters.ToSql()
	if err != nil {
		return nil, err
	}

	results, err := s.queries.GetInboxLogFiltered(ctx, filterBytes)
	if err != nil {
		return nil, err
	}

	// Organize the results by inbox ID
	resultMap := make(map[string][]queries.GetInboxLogFilteredRow)
	for _, result := range results {
		inboxId := result.InboxID
		resultMap[inboxId] = append(resultMap[inboxId], result)
	}

	resps := make([]*identity.GetIdentityUpdatesResponse_Response, len(reqs))
	for i, req := range reqs {
		logEntries := resultMap[req.InboxId]
		updates := make([]*identity.GetIdentityUpdatesResponse_IdentityUpdateLog, len(logEntries))
		for j, entry := range logEntries {
			identity_update := &associations.IdentityUpdate{}
			if err := proto.Unmarshal(entry.IdentityUpdateProto, identity_update); err != nil {
				return nil, err
			}
			updates[j] = &identity.GetIdentityUpdatesResponse_IdentityUpdateLog{
				SequenceId:        uint64(entry.SequenceID),
				ServerTimestampNs: uint64(entry.ServerTimestampNs),
				Update:            identity_update,
			}
		}
		resps[i] = &identity.GetIdentityUpdatesResponse_Response{
			InboxId: req.InboxId,
			Updates: updates,
		}
	}

	return &identity.GetIdentityUpdatesResponse{
		Responses: resps,
	}, nil
}

func (s *ReadStore) FetchKeyPackages(
	ctx context.Context,
	installationIds [][]byte,
) ([]queries.FetchKeyPackagesRow, error) {
	return s.queries.FetchKeyPackages(ctx, installationIds)
}

func (s *ReadStore) QueryGroupMessagesV1(
	ctx context.Context,
	req *mlsv1.QueryGroupMessagesRequest,
) (*mlsv1.QueryGroupMessagesResponse, error) {
	if len(req.GroupId) == 0 {
		return nil, errors.New("group is required")
	}

	sortDesc := true
	var idCursor int64
	var err error
	var messages []queries.GroupMessage
	pageSize := int32(maxPageSize)

	if req.PagingInfo != nil &&
		req.PagingInfo.Direction == mlsv1.SortDirection_SORT_DIRECTION_ASCENDING {
		sortDesc = false
	}

	if req.PagingInfo != nil && req.PagingInfo.Limit > 0 && req.PagingInfo.Limit <= maxPageSize {
		pageSize = int32(req.PagingInfo.Limit)
	}

	if req.PagingInfo != nil && req.PagingInfo.IdCursor != 0 {
		idCursor = int64(req.PagingInfo.IdCursor)
	}

	if idCursor > 0 {
		if sortDesc {
			messages, err = s.queries.QueryGroupMessagesWithCursorDesc(
				ctx,
				queries.QueryGroupMessagesWithCursorDescParams{
					GroupID: req.GroupId,
					Cursor:  idCursor,
					Numrows: pageSize,
				},
			)
		} else {
			messages, err = s.queries.QueryGroupMessagesWithCursorAsc(ctx, queries.QueryGroupMessagesWithCursorAscParams{
				GroupID: req.GroupId,
				Cursor:  idCursor,
				Numrows: pageSize,
			})
		}
	} else {
		messages, err = s.queries.QueryGroupMessages(ctx, queries.QueryGroupMessagesParams{
			GroupID:  req.GroupId,
			Numrows:  pageSize,
			SortDesc: sortDesc,
		})
	}

	if err != nil {
		return nil, err
	}

	out := make([]*mlsv1.GroupMessage, len(messages))
	for idx, msg := range messages {
		out[idx] = &mlsv1.GroupMessage{
			Version: &mlsv1.GroupMessage_V1_{
				V1: &mlsv1.GroupMessage_V1{
					Id:         uint64(msg.ID),
					CreatedNs:  uint64(msg.CreatedAt.UnixNano()),
					GroupId:    msg.GroupID,
					Data:       msg.Data,
					ShouldPush: msg.ShouldPush.Bool,
					SenderHmac: msg.SenderHmac,
					IsCommit:   msg.IsCommit.Bool,
				},
			},
		}
	}

	direction := mlsv1.SortDirection_SORT_DIRECTION_ASCENDING
	if sortDesc {
		direction = mlsv1.SortDirection_SORT_DIRECTION_DESCENDING
	}

	pagingInfo := &mlsv1.PagingInfo{Limit: uint32(pageSize), IdCursor: 0, Direction: direction}
	if len(messages) >= int(pageSize) {
		lastMsg := messages[len(messages)-1]
		pagingInfo.IdCursor = uint64(lastMsg.ID)
	}

	return &mlsv1.QueryGroupMessagesResponse{
		Messages:   out,
		PagingInfo: pagingInfo,
	}, nil
}

func (s *ReadStore) QueryWelcomeMessagesV1(
	ctx context.Context,
	req *mlsv1.QueryWelcomeMessagesRequest,
) (*mlsv1.QueryWelcomeMessagesResponse, error) {
	if len(req.InstallationKey) == 0 {
		return nil, errors.New("installation is required")
	}

	sortDesc := true
	direction := mlsv1.SortDirection_SORT_DIRECTION_DESCENDING
	pageSize := int32(maxPageSize)
	var idCursor int64
	var err error
	var messages []queries.WelcomeMessage

	if req.PagingInfo != nil &&
		req.PagingInfo.Direction == mlsv1.SortDirection_SORT_DIRECTION_ASCENDING {
		sortDesc = false
		direction = mlsv1.SortDirection_SORT_DIRECTION_ASCENDING
	}

	if req.PagingInfo != nil && req.PagingInfo.Limit > 0 && req.PagingInfo.Limit <= maxPageSize {
		pageSize = int32(req.PagingInfo.Limit)
	}

	if req.PagingInfo != nil && req.PagingInfo.IdCursor != 0 {
		idCursor = int64(req.PagingInfo.IdCursor)
	}

	if idCursor > 0 {
		if sortDesc {
			messages, err = s.queries.QueryWelcomeMessagesWithCursorDesc(
				ctx,
				queries.QueryWelcomeMessagesWithCursorDescParams{
					InstallationKey: req.InstallationKey,
					Cursor:          idCursor,
					Numrows:         pageSize,
				},
			)
		} else {
			messages, err = s.queries.QueryWelcomeMessagesWithCursorAsc(ctx, queries.QueryWelcomeMessagesWithCursorAscParams{
				InstallationKey: req.InstallationKey,
				Cursor:          idCursor,
				Numrows:         pageSize,
			})
		}
	} else {
		messages, err = s.queries.QueryWelcomeMessages(ctx, queries.QueryWelcomeMessagesParams{
			InstallationKey: req.InstallationKey,
			Numrows:         pageSize,
			SortDesc:        sortDesc,
		})
	}

	if err != nil {
		return nil, err
	}

	out := make([]*mlsv1.WelcomeMessage, 0, len(messages))
	for _, msg := range messages {
		// Check message type to determine how to construct the proto
		switch msg.MessageType {
		case 0:
			// Regular welcome message
			out = append(out, &mlsv1.WelcomeMessage{
				Version: &mlsv1.WelcomeMessage_V1_{
					V1: &mlsv1.WelcomeMessage_V1{
						Id:              uint64(msg.ID),
						CreatedNs:       uint64(msg.CreatedAt.UnixNano()),
						Data:            msg.Data,
						InstallationKey: msg.InstallationKey,
						HpkePublicKey:   msg.HpkePublicKey,
						WrapperAlgorithm: types.WrapperAlgorithmToProto(
							types.WrapperAlgorithm(msg.WrapperAlgorithm),
						),
						WelcomeMetadata: msg.WelcomeMetadata,
					},
				},
			})
		case 1:
			// Welcome pointer message
			out = append(out, &mlsv1.WelcomeMessage{
				Version: &mlsv1.WelcomeMessage_WelcomePointer_{
					WelcomePointer: &mlsv1.WelcomeMessage_WelcomePointer{
						Id:              uint64(msg.ID),
						CreatedNs:       uint64(msg.CreatedAt.UnixNano()),
						InstallationKey: msg.InstallationKey,
						WelcomePointer:  msg.Data,
						HpkePublicKey:   msg.HpkePublicKey,
						WrapperAlgorithm: types.WrapperAlgorithmToWelcomePointerWrapperAlgorithm(
							types.WrapperAlgorithm(msg.WrapperAlgorithm),
						),
					},
				},
			})
		default:
			// Skip unknown message types
			continue
		}
	}

	pagingInfo := &mlsv1.PagingInfo{Limit: uint32(pageSize), IdCursor: 0, Direction: direction}
	if len(messages) >= int(pageSize) {
		lastMsg := messages[len(messages)-1]
		pagingInfo.IdCursor = uint64(lastMsg.ID)
	}

	return &mlsv1.QueryWelcomeMessagesResponse{
		Messages:   out,
		PagingInfo: pagingInfo,
	}, nil
}

// GroupCatchup names a group and the cursor to resume it from for a batched catch-up.
type GroupCatchup struct {
	GroupID  []byte
	IdCursor uint64
}

// WelcomeCatchup names a welcome topic (by installation) and its resume cursor.
type WelcomeCatchup struct {
	InstallationKey []byte
	IdCursor        uint64
}

// clampCursor narrows a uint64 client cursor to the int64 the SQL columns use,
// saturating at MaxInt64. A cursor above 2^63-1 can't match any real row id, so
// it must return no rows; without the clamp the int64 conversion would wrap to a
// negative value and the `id > cursor` predicate would replay the entire history.
func clampCursor(c uint64) int64 {
	const maxInt64 = 1<<63 - 1
	if c > maxInt64 {
		return maxInt64
	}
	return int64(c)
}

const queryGroupMessagesBatch = `
SELECT m.id, m.created_at, m.group_id, m.data, m.is_commit, m.sender_hmac, m.should_push
FROM unnest($1::BYTEA[], $2::BIGINT[]) AS f (group_id, id_cursor)
CROSS JOIN LATERAL (
	SELECT g.id, g.created_at, g.group_id, g.data, g.is_commit, g.sender_hmac, g.should_push
	FROM group_messages g
	WHERE g.group_id = f.group_id AND g.id > f.id_cursor
	ORDER BY g.id ASC
	LIMIT $3
) m
ORDER BY m.group_id ASC, m.id ASC`

// QueryGroupMessagesBatch fetches catch-up messages for many groups in one round-trip
// (XIP-83). Each group resumes from its own cursor and is capped at perGroupLimit rows;
// a group returning fewer than perGroupLimit rows is fully caught up. Ordered by
// (group_id, id). The composite index (group_id, id) makes each per-group seek cheap.
//
// Filter keys MUST be unique: unnest preserves duplicates and CROSS JOIN LATERAL runs the
// subquery once per occurrence, so a repeated key returns its rows more than once and
// breaks the caller's per-key pagination count. The Subscribe handler deduplicates adds
// before calling this.
func (s *ReadStore) QueryGroupMessagesBatch(
	ctx context.Context,
	filters []GroupCatchup,
	perGroupLimit int32,
) ([]*mlsv1.GroupMessage, error) {
	if len(filters) == 0 {
		return nil, nil
	}
	groupIds := make([][]byte, len(filters))
	cursors := make([]int64, len(filters))
	for i, f := range filters {
		groupIds[i] = f.GroupID
		cursors[i] = clampCursor(f.IdCursor)
	}
	rows, err := s.db.QueryContext(
		ctx,
		queryGroupMessagesBatch,
		pq.Array(groupIds),
		pq.Array(cursors),
		perGroupLimit,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []*mlsv1.GroupMessage
	for rows.Next() {
		var (
			id         int64
			createdAt  time.Time
			groupID    []byte
			data       []byte
			isCommit   sql.NullBool
			senderHmac []byte
			shouldPush sql.NullBool
		)
		if err := rows.Scan(&id, &createdAt, &groupID, &data, &isCommit, &senderHmac, &shouldPush); err != nil {
			return nil, err
		}
		out = append(out, &mlsv1.GroupMessage{
			Version: &mlsv1.GroupMessage_V1_{
				V1: &mlsv1.GroupMessage_V1{
					Id:         uint64(id),
					CreatedNs:  uint64(createdAt.UnixNano()),
					GroupId:    groupID,
					Data:       data,
					ShouldPush: shouldPush.Bool,
					SenderHmac: senderHmac,
					IsCommit:   isCommit.Bool,
				},
			},
		})
	}
	return out, rows.Err()
}

const queryWelcomeMessagesBatch = `
SELECT m.id, m.created_at, m.installation_key, m.data, m.hpke_public_key, m.wrapper_algorithm, m.welcome_metadata, m.message_type
FROM unnest($1::BYTEA[], $2::BIGINT[]) AS f (installation_key, id_cursor)
CROSS JOIN LATERAL (
	SELECT w.id, w.created_at, w.installation_key, w.data, w.hpke_public_key, w.wrapper_algorithm, w.welcome_metadata, w.message_type
	FROM welcome_messages w
	WHERE w.installation_key = f.installation_key AND w.id > f.id_cursor
	ORDER BY w.id ASC
	LIMIT $3
) m
ORDER BY m.installation_key ASC, m.id ASC`

// QueryWelcomeMessagesBatch is the welcome-topic analogue of QueryGroupMessagesBatch.
func (s *ReadStore) QueryWelcomeMessagesBatch(
	ctx context.Context,
	filters []WelcomeCatchup,
	perGroupLimit int32,
) ([]*mlsv1.WelcomeMessage, error) {
	if len(filters) == 0 {
		return nil, nil
	}
	keys := make([][]byte, len(filters))
	cursors := make([]int64, len(filters))
	for i, f := range filters {
		keys[i] = f.InstallationKey
		cursors[i] = clampCursor(f.IdCursor)
	}
	rows, err := s.db.QueryContext(
		ctx,
		queryWelcomeMessagesBatch,
		pq.Array(keys),
		pq.Array(cursors),
		perGroupLimit,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []*mlsv1.WelcomeMessage
	for rows.Next() {
		var (
			id               int64
			createdAt        time.Time
			installationKey  []byte
			data             []byte
			hpkePublicKey    []byte
			wrapperAlgorithm int16
			welcomeMetadata  []byte
			messageType      int16
		)
		if err := rows.Scan(&id, &createdAt, &installationKey, &data, &hpkePublicKey, &wrapperAlgorithm, &welcomeMetadata, &messageType); err != nil {
			return nil, err
		}
		switch messageType {
		case 0:
			out = append(out, &mlsv1.WelcomeMessage{
				Version: &mlsv1.WelcomeMessage_V1_{
					V1: &mlsv1.WelcomeMessage_V1{
						Id:              uint64(id),
						CreatedNs:       uint64(createdAt.UnixNano()),
						Data:            data,
						InstallationKey: installationKey,
						HpkePublicKey:   hpkePublicKey,
						WrapperAlgorithm: types.WrapperAlgorithmToProto(
							types.WrapperAlgorithm(wrapperAlgorithm),
						),
						WelcomeMetadata: welcomeMetadata,
					},
				},
			})
		case 1:
			out = append(out, &mlsv1.WelcomeMessage{
				Version: &mlsv1.WelcomeMessage_WelcomePointer_{
					WelcomePointer: &mlsv1.WelcomeMessage_WelcomePointer{
						Id:              uint64(id),
						CreatedNs:       uint64(createdAt.UnixNano()),
						InstallationKey: installationKey,
						WelcomePointer:  data,
						HpkePublicKey:   hpkePublicKey,
						WrapperAlgorithm: types.WrapperAlgorithmToWelcomePointerWrapperAlgorithm(
							types.WrapperAlgorithm(wrapperAlgorithm),
						),
					},
				},
			})
		default:
			// Parity with the live dispatch path (worker.go): surface, don't swallow.
			s.log.Error("unknown welcome message type", zap.Int16("message_type", messageType))
		}
	}
	return out, rows.Err()
}

func (s *ReadStore) QueryCommitLog(
	ctx context.Context,
	req *mlsv1.QueryCommitLogRequest,
) (*mlsv1.QueryCommitLogResponse, error) {
	if len(req.GetGroupId()) == 0 {
		return nil, errors.New("group is required")
	}
	if req.GetPagingInfo().GetDirection() == mlsv1.SortDirection_SORT_DIRECTION_DESCENDING {
		return nil, errors.New("descending direction is not supported")
	}
	pageSize := int32(req.GetPagingInfo().GetLimit())
	if pageSize <= 0 || pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	idCursor := int64(req.GetPagingInfo().GetIdCursor())

	entries, err := s.queries.QueryCommitLogV2(ctx, queries.QueryCommitLogV2Params{
		GroupID: req.GroupId,
		Cursor:  idCursor,
		Numrows: pageSize,
	})
	if err != nil {
		return nil, err
	}

	out := make([]*message_contents.CommitLogEntry, len(entries))
	for idx, entry := range entries {
		signature := &associations.RecoverableEd25519Signature{}
		err := proto.Unmarshal(entry.SerializedSignature, signature)
		if err != nil {
			return nil, err
		}
		out[idx] = &message_contents.CommitLogEntry{
			SequenceId:               uint64(entry.ID),
			SerializedCommitLogEntry: entry.SerializedEntry,
			Signature:                signature,
		}
	}

	pagingInfo := &mlsv1.PagingInfo{
		Limit:     uint32(pageSize),
		IdCursor:  0,
		Direction: mlsv1.SortDirection_SORT_DIRECTION_ASCENDING,
	}
	// It is strange behavior, but we must follow the semantics of other queries in v3 - only setting the IdCursor
	// to a non-zero value if we have a full page of results
	if len(entries) >= int(pageSize) {
		lastEntry := entries[len(entries)-1]
		pagingInfo.IdCursor = uint64(lastEntry.ID)
	}

	return &mlsv1.QueryCommitLogResponse{
		GroupId:          req.GroupId,
		CommitLogEntries: out,
		PagingInfo:       pagingInfo,
	}, nil
}

// Returns an array of the same length as the input groupIds in the original order. GroupIDs with no messages are represented by a nil value
func (s *ReadStore) GetNewestGroupMessage(
	ctx context.Context,
	groupIds [][]byte,
) ([]*queries.GetNewestGroupMessageRow, error) {
	results, err := s.queries.GetNewestGroupMessage(ctx, groupIds)
	if err != nil {
		return nil, err
	}

	// Create a map for quick lookup of results by group ID
	resultMap := make(map[string]*queries.GetNewestGroupMessageRow)
	for i, result := range results {
		key := string(result.GroupID)
		resultMap[key] = &results[i]
	}

	// Build the output array in the same order as input groupIds
	output := make([]*queries.GetNewestGroupMessageRow, len(groupIds))
	for i, groupId := range groupIds {
		if result, exists := resultMap[string(groupId)]; exists {
			output[i] = result
		}
	}

	return output, nil
}

func (s *ReadStore) GetNewestGroupMessageMetadata(
	ctx context.Context,
	groupIds [][]byte,
) ([]*queries.GetNewestGroupMessageMetadataRow, error) {
	results, err := s.queries.GetNewestGroupMessageMetadata(ctx, groupIds)
	if err != nil {
		return nil, err
	}

	// Create a map for quick lookup of results by group ID
	resultMap := make(map[string]*queries.GetNewestGroupMessageMetadataRow)
	for i, result := range results {
		key := string(result.GroupID)
		resultMap[key] = &results[i]
	}

	// Build the output array in the same order as input groupIds
	output := make([]*queries.GetNewestGroupMessageMetadataRow, len(groupIds))
	for i, groupId := range groupIds {
		if result, exists := resultMap[string(groupId)]; exists {
			output[i] = result
		}
	}

	return output, nil
}
