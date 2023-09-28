package metrics

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

var StoredMessages = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "xmtp_stored_messages",
		Help: "Count of stored messages",
	},
)

func EmitStoredMessages(ctx context.Context, db *sql.DB, logger *zap.Logger) {
	count, err := messageCountEstimate(db)
	if err != nil {
		logger.Error("counting messages", zap.Error(err))
	}
	StoredMessages.Set(float64(count))
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
