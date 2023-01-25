package metrics

import (
	"context"
	"database/sql"
	"fmt"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.uber.org/zap"
)

var StoredMessages = stats.Int64("stored_messages", "Count of stored messages", stats.UnitDimensionless)
var StoredMessageView = &view.View{
	Name:        "xmtp_stored_messages",
	Measure:     StoredMessages,
	Description: "Count of stored messages",
	Aggregation: view.LastValue(),
}

func EmitStoredMessages(ctx context.Context, db *sql.DB, logger *zap.Logger) {
	count, err := messageCountEstimate(db)
	if err != nil {
		logger.Error("counting messages", zap.Error(err))
	}
	stats.Record(ctx, StoredMessages.M(count))
}

func messageCountEstimate(db *sql.DB) (count int64, err error) {
	rows, err := db.Query("SELECT reltuples::bigint AS estimate FROM pg_class WHERE oid = 'public.message'::regclass")
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	if !rows.Next() {
		return 0, fmt.Errorf("count has no rows")
	}
	err = rows.Scan(&count)
	return count, err
}
