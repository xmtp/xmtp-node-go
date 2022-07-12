package store

import (
	"database/sql"

	sqlBuilder "github.com/huandu/go-sqlbuilder"
	"github.com/status-im/go-waku/waku/persistence"
	"github.com/status-im/go-waku/waku/v2/protocol"
	"github.com/status-im/go-waku/waku/v2/protocol/pb"
)

/**
Executes the appropriate database queries to find messages and build a response based on a HistoryQuery
**/
func FindMessages(db *sql.DB, query *pb.HistoryQuery) (res *pb.HistoryResponse, err error) {
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

func getContentTopics(filters []*pb.ContentFilter) []string {
	out := make([]string, len(filters))
	for i, filter := range filters {
		out[i] = filter.ContentTopic
	}
	return out
}

func buildSqlQuery(query *pb.HistoryQuery) (querySql string, args []interface{}, err error) {
	sb := sqlBuilder.PostgreSQL.NewSelectBuilder()

	sb.Select("*").From("message")

	contentTopics := getContentTopics(query.ContentFilters)
	if len(contentTopics) > 0 {
		sb.Where(sb.In("contentTopic", sqlBuilder.Flatten(contentTopics)...))
	}

	if query.PubsubTopic != "" {
		sb.Where(sb.Equal("pubsubTopic", query.PubsubTopic))
	}

	if query.StartTime != 0 {
		sb.Where(sb.GreaterEqualThan("senderTimestamp", query.StartTime))
	}

	if query.EndTime != 0 {
		sb.Where(sb.LessEqualThan("senderTimestamp", query.EndTime))
	}

	pagingInfo := query.PagingInfo
	addPagination(sb, pagingInfo)

	direction := getDirection(pagingInfo)
	addSort(sb, direction)

	querySql, args = sb.Build()

	return querySql, args, err
}

func addPagination(sb *sqlBuilder.SelectBuilder, pagination *pb.PagingInfo) {
	if pagination == nil {
		pagination = new(pb.PagingInfo)
	}

	// Set page size
	pageSize := pagination.PageSize
	addLimit(sb, pageSize)

	cursor := pagination.Cursor
	direction := pagination.Direction
	if cursor != nil {
		addCursor(sb, cursor, direction)
	}
}

func getDirection(pagingInfo *pb.PagingInfo) (direction pb.PagingInfo_Direction) {
	if pagingInfo == nil {
		direction = pb.PagingInfo_BACKWARD
	} else {
		direction = pagingInfo.Direction
	}
	return
}

func addLimit(sb *sqlBuilder.SelectBuilder, pageSize uint64) {
	if pageSize == 0 {
		pageSize = maxPageSize
	}
	sb.Limit(int(pageSize))
}

func addSort(sb *sqlBuilder.SelectBuilder, direction pb.PagingInfo_Direction) {
	switch direction {
	case pb.PagingInfo_BACKWARD:
		sb.OrderBy("senderTimestamp desc", "id desc", "pubsubTopic desc", "receiverTimestamp desc")
	case pb.PagingInfo_FORWARD:
		sb.OrderBy("senderTimestamp asc", "id asc", "pubsubTopic asc", "receiverTimestamp asc")
	}

}

func addCursor(sb *sqlBuilder.SelectBuilder, cursor *pb.Index, direction pb.PagingInfo_Direction) {
	switch direction {
	case pb.PagingInfo_FORWARD:
		if cursor.SenderTime != 0 && cursor.Digest != nil {
			// This is tricky. The current implementation does a complex sort by senderTimestamp, digest (id), pubsub topic, and receiverTimestamp
			// This is also used for cursor based pagination
			// I am going for 1:1 parity right now, and not worried about performance.
			// But this, and the sort, is going to be a real performance issue without indexing.
			// Alternatively, I could use a derived table/CTE to accomplish this
			sb.Where(
				sb.Or(
					sb.GreaterThan("senderTimestamp", cursor.SenderTime),
					sb.And(sb.Equal("senderTimestamp", cursor.SenderTime), sb.GreaterThan("id", cursor.Digest)),
					// sb.And(sb.Equal("senderTimestamp", cursor.SenderTime), sb.Equal("id", cursor.Digest), sb.GreaterThan("pubsubTopic", cursor.PubsubTopic)),
				),
			)
		} else if cursor.Digest != nil {
			sb.Where(
				sb.Or(
					sb.GreaterThan("id", cursor.Digest),
					// sb.And(sb.Equal("id", cursor.Digest), sb.GreaterThan("pubsubTopic", cursor.PubsubTopic)),
				),
			)
		}
	case pb.PagingInfo_BACKWARD:
		if cursor.SenderTime != 0 && cursor.Digest != nil {
			sb.Where(
				sb.Or(
					sb.LessThan("senderTimestamp", cursor.SenderTime),
					sb.And(sb.Equal("senderTimestamp", cursor.SenderTime), sb.LessThan("id", cursor.Digest)),
					// sb.And(sb.Equal("senderTimestamp", cursor.SenderTime), sb.Equal("id", cursor.Digest), sb.LessThan("pubsubTopic", cursor.PubsubTopic)),
				),
			)
		} else if cursor.Digest != nil {
			sb.Where(
				sb.Or(
					sb.LessThan("id", cursor.Digest),
					// sb.And(sb.Equal("id", cursor.Digest), sb.LessThan("pubsubTopic", cursor.PubsubTopic)),
				),
			)
		}
	}
}

func buildResponse(rows *sql.Rows, query *pb.HistoryQuery) (*pb.HistoryResponse, error) {
	storedMessages, err := rowsToMessages(rows)
	if err != nil {
		return nil, err
	}
	messages := messagesFromStoredMessages(storedMessages)
	pagingInfo, err := buildPagingInfo(storedMessages, query.PagingInfo)
	if err != nil {
		return nil, err
	}

	return &pb.HistoryResponse{
		Messages:   messages,
		PagingInfo: pagingInfo,
	}, nil
}

// Builds the paging info based on the response.
// Since we do not get a total count in our query, and I would like to avoid adding that, there is one important difference with go-waku here.
// On the last page of results, go-waku will reduce the page-size to the number of remaining items.
// We will leave it intact, as we do not have that information on hand.
// Clients may be relying on this behaviour to know when to stop paginating
func buildPagingInfo(messages []persistence.StoredMessage, pagingInfo *pb.PagingInfo) (*pb.PagingInfo, error) {
	currentPageSize := getPageSize(pagingInfo)
	newPageSize := minOf(currentPageSize, maxPageSize)
	direction := getDirection(pagingInfo)

	if len(messages) < currentPageSize {
		return &pb.PagingInfo{PageSize: uint64(0), Cursor: nil, Direction: direction}, nil
	}

	newCursor, err := findNextCursor(messages)
	if err != nil {
		return nil, err
	}

	return &pb.PagingInfo{
		PageSize:  uint64(newPageSize),
		Cursor:    newCursor,
		Direction: direction,
	}, nil
}

// func getCursor(pagingInfo *pb.PagingInfo) *pb.Index {
// 	if pagingInfo == nil {
// 		return nil
// 	}
// 	return pagingInfo.Cursor
// }

func findNextCursor(messages []persistence.StoredMessage) (*pb.Index, error) {
	if len(messages) == 0 {
		return nil, nil
	}

	lastMessage := messages[len(messages)-1]
	envelope := protocol.NewEnvelope(lastMessage.Message, lastMessage.ReceiverTime, lastMessage.PubsubTopic)
	return computeIndex(envelope)
}

func getPageSize(pagingInfo *pb.PagingInfo) int {
	if pagingInfo == nil || pagingInfo.PageSize == 0 {
		return maxPageSize
	}
	return int(pagingInfo.PageSize)
}

func messagesFromStoredMessages(storedMessages []persistence.StoredMessage) []*pb.WakuMessage {
	out := make([]*pb.WakuMessage, len(storedMessages))
	for i, msg := range storedMessages {
		out[i] = msg.Message
	}
	return out
}

func rowsToMessages(rows *sql.Rows) (result []persistence.StoredMessage, err error) {
	defer rows.Close()

	for rows.Next() {
		var id []byte
		var receiverTimestamp int64
		var senderTimestamp int64
		var contentTopic string
		var payload []byte
		var version uint32
		var pubsubTopic string

		err = rows.Scan(&id, &receiverTimestamp, &senderTimestamp, &contentTopic, &pubsubTopic, &payload, &version)
		if err != nil {
			return
		}

		msg := new(pb.WakuMessage)
		msg.ContentTopic = contentTopic
		msg.Payload = payload
		msg.Timestamp = senderTimestamp
		msg.Version = version

		record := persistence.StoredMessage{
			ID:           id,
			PubsubTopic:  pubsubTopic,
			ReceiverTime: receiverTimestamp,
			Message:      msg,
		}

		result = append(result, record)
	}

	err = rows.Err()
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
