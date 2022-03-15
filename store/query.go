package store

import (
	"database/sql"

	sqlBuilder "github.com/huandu/go-sqlbuilder"
	"github.com/jmoiron/sqlx"
	"github.com/status-im/go-waku/waku/persistence"
	"github.com/status-im/go-waku/waku/v2/protocol/pb"
)

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
	sb := sqlBuilder.SQLite.NewSelectBuilder()

	sb.Select("*").From("message")

	contentTopics := getContentTopics(query.ContentFilters)
	if len(contentTopics) > 0 {
		sb.Where(sb.In("contentTopic", contentTopics))
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

	sb.Limit(10)

	addPagination(sb, query.PagingInfo)

	querySql, args = sb.Build()
	// Use SQLX to convert IN statements to multiple arguments. Only way I could find to make IN statements work with SQLite
	querySql, args, err = sqlx.In(querySql, args...)

	return
}

func addPagination(sb *sqlBuilder.SelectBuilder, pagination *pb.PagingInfo) *sqlBuilder.SelectBuilder {
	// Not yet implemented
	return sb
}

func buildResponse(rows *sql.Rows, query *pb.HistoryQuery) (*pb.HistoryResponse, error) {
	storedMessages, err := rowsToMessages(rows)
	return &pb.HistoryResponse{
		Messages:   messagesFromStoredMessages(storedMessages),
		PagingInfo: nil,
	}, err
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
