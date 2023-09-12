package store

import (
	"crypto/sha256"
	"database/sql"

	sqlBuilder "github.com/huandu/go-sqlbuilder"
	messagev1 "github.com/xmtp/proto/v3/go/message_api/v1"
)

func (s *Store) Query(query *messagev1.QueryRequest) (res *messagev1.QueryResponse, err error) {
	sql, args := buildSqlQuery(query)

	rows, err := s.config.DB.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return buildResponse(rows, query)
}

func buildSqlQuery(query *messagev1.QueryRequest) (querySql string, args []interface{}) {
	pagingInfo := query.PagingInfo
	if pagingInfo == nil {
		pagingInfo = new(messagev1.PagingInfo)
	}
	cursor := pagingInfo.Cursor
	direction := getDirection(pagingInfo)

	if cursor == nil {
		sb := buildSqlQueryWithoutCursor(query)
		return sb.Build()
	}

	ub := sqlBuilder.PostgreSQL.NewUnionBuilder()

	sb1 := buildSqlQueryWithoutCursor(query)
	index := cursor.GetIndex()
	switch direction {
	case messagev1.SortDirection_SORT_DIRECTION_ASCENDING:
		if index.SenderTimeNs != 0 && index.Digest != nil {
			sb1.Where(sb1.GreaterThan("senderTimestamp", index.SenderTimeNs))
		} else if index.Digest != nil {
			sb1.Where(sb1.GreaterThan("id", index.Digest))
		}
	case messagev1.SortDirection_SORT_DIRECTION_DESCENDING:
		if index.SenderTimeNs != 0 && index.Digest != nil {
			sb1.Where(sb1.LessThan("senderTimestamp", index.SenderTimeNs))
		} else if index.Digest != nil {
			sb1.Where(sb1.LessThan("id", index.Digest))
		}
	}

	sb2 := buildSqlQueryWithoutCursor(query)
	switch direction {
	case messagev1.SortDirection_SORT_DIRECTION_ASCENDING:
		if index.SenderTimeNs != 0 && index.Digest != nil {
			sb2.Where(sb2.And(sb2.Equal("senderTimestamp", index.SenderTimeNs), sb2.GreaterThan("id", index.Digest)))
		} else if index.Digest != nil {
			sb2.Where(sb2.GreaterThan("id", index.Digest))
		}
	case messagev1.SortDirection_SORT_DIRECTION_DESCENDING:
		if index.SenderTimeNs != 0 && index.Digest != nil {
			sb2.Where(sb2.And(sb2.Equal("senderTimestamp", index.SenderTimeNs), sb2.LessThan("id", index.Digest)))
		} else if index.Digest != nil {
			sb2.Where(sb2.LessThan("id", index.Digest))
		}
	}

	ub.Union(sb1, sb2)

	// Add limit.
	ub.Limit(getPageSize(pagingInfo))

	// Add sorting.
	switch direction {
	case messagev1.SortDirection_SORT_DIRECTION_DESCENDING:
		ub.OrderBy("senderTimestamp desc", "id desc")
	case messagev1.SortDirection_SORT_DIRECTION_ASCENDING:
		ub.OrderBy("senderTimestamp asc", "id asc")
	}

	return ub.Build()
}

func buildSqlQueryWithoutCursor(query *messagev1.QueryRequest) *sqlBuilder.SelectBuilder {
	sb := sqlBuilder.PostgreSQL.NewSelectBuilder()

	sb.Select("id, receivertimestamp, sendertimestamp, contenttopic, pubsubtopic, payload, version").From("message")

	if len(query.ContentTopics) > 0 {
		sb.Where(sb.In("contentTopic", sqlBuilder.Flatten(query.ContentTopics)...))
	}

	if query.StartTimeNs != 0 {
		sb.Where(sb.GreaterEqualThan("senderTimestamp", query.StartTimeNs))
	}

	if query.EndTimeNs != 0 {
		sb.Where(sb.LessEqualThan("senderTimestamp", query.EndTimeNs))
	}

	pagingInfo := query.PagingInfo
	if pagingInfo == nil {
		pagingInfo = new(messagev1.PagingInfo)
	}

	pageSize := pagingInfo.Limit
	addLimit(sb, pageSize)

	direction := getDirection(pagingInfo)
	addSort(sb, direction)

	return sb
}

func getDirection(pagingInfo *messagev1.PagingInfo) (direction messagev1.SortDirection) {
	if pagingInfo == nil || pagingInfo.Direction == messagev1.SortDirection_SORT_DIRECTION_UNSPECIFIED {
		direction = messagev1.SortDirection_SORT_DIRECTION_DESCENDING
	} else {
		direction = pagingInfo.Direction
	}
	return
}

func addLimit(sb *sqlBuilder.SelectBuilder, pageSize uint32) {
	if pageSize == 0 {
		pageSize = maxPageSize
	}
	sb.Limit(int(pageSize))
}

func addSort(sb *sqlBuilder.SelectBuilder, direction messagev1.SortDirection) {
	switch direction {
	case messagev1.SortDirection_SORT_DIRECTION_DESCENDING:
		sb.OrderBy("senderTimestamp desc", "id desc")
	case messagev1.SortDirection_SORT_DIRECTION_ASCENDING:
		sb.OrderBy("senderTimestamp asc", "id asc")
	}
}

func buildResponse(rows *sql.Rows, query *messagev1.QueryRequest) (*messagev1.QueryResponse, error) {
	envs, err := rowsToEnvelopes(rows)
	if err != nil {
		return nil, err
	}
	pagingInfo, err := buildPagingInfo(envs, query.PagingInfo)
	if err != nil {
		return nil, err
	}

	return &messagev1.QueryResponse{
		Envelopes:  envs,
		PagingInfo: pagingInfo,
	}, nil
}

// Builds the paging info based on the response.
// Clients may be relying on this behaviour to know when to stop paginating
func buildPagingInfo(envs []*messagev1.Envelope, pagingInfo *messagev1.PagingInfo) (*messagev1.PagingInfo, error) {
	currentPageSize := getPageSize(pagingInfo)
	newPageSize := minOf(currentPageSize, maxPageSize)
	direction := getDirection(pagingInfo)

	if len(envs) < currentPageSize {
		return &messagev1.PagingInfo{Limit: 0, Cursor: nil, Direction: direction}, nil
	}

	newCursor, err := findNextCursor(envs)
	if err != nil {
		return nil, err
	}

	return &messagev1.PagingInfo{
		Limit:     uint32(newPageSize),
		Cursor:    newCursor,
		Direction: direction,
	}, nil
}

func findNextCursor(envs []*messagev1.Envelope) (*messagev1.Cursor, error) {
	if len(envs) == 0 {
		return nil, nil
	}

	lastEnv := envs[len(envs)-1]
	return buildCursor(lastEnv), nil
}

func buildCursor(env *messagev1.Envelope) *messagev1.Cursor {
	return &messagev1.Cursor{
		Cursor: &messagev1.Cursor_Index{
			Index: &messagev1.IndexCursor{
				Digest:       computeDigest(env),
				SenderTimeNs: env.TimestampNs,
			},
		},
	}
}

func computeDigest(env *messagev1.Envelope) []byte {
	digest := sha256.Sum256(append([]byte(env.ContentTopic), env.Message...))
	return digest[:]
}

func getPageSize(pagingInfo *messagev1.PagingInfo) int {
	if pagingInfo == nil || pagingInfo.Limit == 0 {
		return maxPageSize
	}
	return int(pagingInfo.Limit)
}

func rowsToEnvelopes(rows *sql.Rows) ([]*messagev1.Envelope, error) {
	defer rows.Close()

	var result []*messagev1.Envelope
	for rows.Next() {
		var id []byte
		var receiverTimestamp int64
		var version uint32
		var pubsubTopic string
		var env messagev1.Envelope

		err := rows.Scan(&id, &receiverTimestamp, &env.TimestampNs, &env.ContentTopic, &pubsubTopic, &env.Message, &version)
		if err != nil {
			return nil, err
		}

		result = append(result, &env)
	}

	err := rows.Err()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func minOf(vars ...int) int {
	min := vars[0]

	for _, i := range vars {
		if min > i {
			min = i
		}
	}

	return min
}
