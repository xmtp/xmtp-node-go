package store

import (
	"database/sql"

	sqlBuilder "github.com/huandu/go-sqlbuilder"
	messagev1 "github.com/xmtp/proto/v3/go/message_api/v1"
)

/*
*
Executes the appropriate database queries to find messages and build a response based on a HistoryQuery
*
*/
func FindMessages(db *sql.DB, query *messagev1.QueryRequest) (res *messagev1.QueryResponse, err error) {
	var rows *sql.Rows
	sql, args, err := buildSqlQuery(query)
	if err != nil {
		return
	}

	rows, err = db.Query(sql, args...)
	if err != nil {
		return
	}

	res, err = buildResponse(rows, query)

	return
}

func buildSqlQuery(query *messagev1.QueryRequest) (querySql string, args []interface{}, err error) {
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
	addPagination(sb, pagingInfo)

	direction := getDirection(pagingInfo)
	addSort(sb, direction)

	querySql, args = sb.Build()

	return querySql, args, err
}

func addPagination(sb *sqlBuilder.SelectBuilder, pagination *messagev1.PagingInfo) {
	if pagination == nil {
		pagination = new(messagev1.PagingInfo)
	}

	// Set page size
	pageSize := pagination.Limit
	addLimit(sb, pageSize)

	cursor := pagination.Cursor
	direction := pagination.Direction
	if cursor != nil {
		addCursor(sb, cursor, direction)
	}
}

func getDirection(pagingInfo *messagev1.PagingInfo) (direction messagev1.SortDirection) {
	if pagingInfo == nil {
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
		sb.OrderBy("senderTimestamp desc", "id desc", "pubsubTopic desc", "receiverTimestamp desc")
	case messagev1.SortDirection_SORT_DIRECTION_ASCENDING:
		sb.OrderBy("senderTimestamp asc", "id asc", "pubsubTopic asc", "receiverTimestamp asc")
	}

}

func addCursor(sb *sqlBuilder.SelectBuilder, cursor *messagev1.Cursor, direction messagev1.SortDirection) {
	index := cursor.GetIndex()
	switch direction {
	case messagev1.SortDirection_SORT_DIRECTION_ASCENDING:
		if index.SenderTimeNs != 0 && index.Digest != nil {
			// This is tricky. The current implementation does a complex sort by senderTimestamp, digest (id), pubsub topic, and receiverTimestamp
			// This is also used for cursor based pagination
			// I am going for 1:1 parity right now, and not worried about performance.
			// But this, and the sort, is going to be a real performance issue without indexing.
			// Alternatively, I could use a derived table/CTE to accomplish this
			sb.Where(
				sb.Or(
					sb.GreaterThan("senderTimestamp", index.SenderTimeNs),
					sb.And(sb.Equal("senderTimestamp", index.SenderTimeNs), sb.GreaterThan("id", index.Digest)),
					// sb.And(sb.Equal("senderTimestamp", cursor.SenderTime), sb.Equal("id", cursor.Digest), sb.GreaterThan("pubsubTopic", cursor.PubsubTopic)),
				),
			)
		} else if index.Digest != nil {
			sb.Where(
				sb.Or(
					sb.GreaterThan("id", index.Digest),
					// sb.And(sb.Equal("id", cursor.Digest), sb.GreaterThan("pubsubTopic", cursor.PubsubTopic)),
				),
			)
		}
	case messagev1.SortDirection_SORT_DIRECTION_DESCENDING:
		if index.SenderTimeNs != 0 && index.Digest != nil {
			sb.Where(
				sb.Or(
					sb.LessThan("senderTimestamp", index.SenderTimeNs),
					sb.And(sb.Equal("senderTimestamp", index.SenderTimeNs), sb.LessThan("id", index.Digest)),
					// sb.And(sb.Equal("senderTimestamp", cursor.SenderTime), sb.Equal("id", cursor.Digest), sb.LessThan("pubsubTopic", cursor.PubsubTopic)),
				),
			)
		} else if index.Digest != nil {
			sb.Where(
				sb.Or(
					sb.LessThan("id", index.Digest),
					// sb.And(sb.Equal("id", cursor.Digest), sb.LessThan("pubsubTopic", cursor.PubsubTopic)),
				),
			)
		}
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
