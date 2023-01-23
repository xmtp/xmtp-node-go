package store

import (
	"context"
	"database/sql"
	"strings"

	"github.com/status-im/go-waku/waku/persistence"
	"github.com/status-im/go-waku/waku/v2/protocol"
	"github.com/status-im/go-waku/waku/v2/protocol/pb"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/migrate"
	"github.com/xmtp/xmtp-node-go/pkg/migrations/messages"
	"go.uber.org/zap"
)

// DBStore is a MessageProvider that has a *sql.DB connection
type DBStore struct {
	persistence.MessageProvider
	db  *sql.DB
	log *zap.Logger
}

// DBOption is an optional setting that can be used to configure the DBStore
type DBOption func(*DBStore) error

// WithDBStoreDB is a DBOption that lets you use any custom *sql.DB with a DBStore.
func WithDBStoreDB(db *sql.DB) DBOption {
	return func(d *DBStore) error {
		d.db = db
		return nil
	}
}

// Creates a new DB store using the db specified via options.
// It will create a messages table if it does not exist and
// clean up records according to the retention policy used
func NewDBStore(log *zap.Logger, options ...DBOption) (*DBStore, error) {
	result := new(DBStore)
	result.log = log.Named("dbstore")

	for _, opt := range options {
		err := opt(result)
		if err != nil {
			return nil, err
		}
	}

	err := result.migrate()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (d *DBStore) migrate() error {
	ctx := context.Background()
	db := bun.NewDB(d.db, pgdialect.New())
	migrator := migrate.NewMigrator(db, messages.Migrations)
	err := migrator.Init(ctx)
	if err != nil {
		return err
	}

	group, err := migrator.Migrate(ctx)
	if err != nil {
		return err
	}
	if group.IsZero() {
		d.log.Info("No new migrations to run")
	}

	return nil
}

// Closes a DB connection
func (d *DBStore) Stop() {
	d.db.Close()
}

// Inserts a WakuMessage into the DB
func (d *DBStore) Put(env *protocol.Envelope) error {
	cursor := env.Index()
	pubsubTopic := env.PubsubTopic()
	message := env.Message()
	shouldExpire := !isXMTP(message)
	stmt, err := d.db.Prepare("INSERT INTO message (id, receiverTimestamp, senderTimestamp, contentTopic, pubsubTopic, payload, version, should_expire) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(cursor.Digest, cursor.ReceiverTime, message.Timestamp, message.ContentTopic, pubsubTopic, message.Payload, message.Version, shouldExpire)
	if err != nil {
		return err
	}

	return nil
}

func isXMTP(msg *pb.WakuMessage) bool {
	return strings.HasPrefix(msg.ContentTopic, "/xmtp/")
}
