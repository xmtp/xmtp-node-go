package store

import (
	"database/sql"
	"time"

	"github.com/status-im/go-waku/waku/persistence"
	"github.com/status-im/go-waku/waku/v2/protocol/pb"
	"go.uber.org/zap"
)

// DBStore is a MessageProvider that has a *sql.DB connection
type DBStore struct {
	persistence.MessageProvider
	db  *sql.DB
	log *zap.SugaredLogger

	maxMessages int
	maxDuration time.Duration
}

// DBOption is an optional setting that can be used to configure the DBStore
type DBOption func(*DBStore) error

// WithDB is a DBOption that lets you use any custom *sql.DB with a DBStore.
func WithDB(db *sql.DB) DBOption {
	return func(d *DBStore) error {
		d.db = db
		return nil
	}
}

// Creates a new DB store using the db specified via options.
// It will create a messages table if it does not exist and
// clean up records according to the retention policy used
func NewDBStore(log *zap.SugaredLogger, options ...DBOption) (*DBStore, error) {
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
	sqlStmt := `CREATE TABLE IF NOT EXISTS message (
		id BLOB,
		receiverTimestamp INTEGER NOT NULL,
		senderTimestamp INTEGER NOT NULL,
		contentTopic BLOB NOT NULL,
		pubsubTopic BLOB NOT NULL,
		payload BLOB,
		version INTEGER NOT NULL DEFAULT 0,
		CONSTRAINT messageIndex PRIMARY KEY (senderTimestamp, id, pubsubTopic)
	) WITHOUT ROWID;`
	_, err := d.db.Exec(sqlStmt)
	if err != nil {
		return err
	}
	return nil
}

// Closes a DB connection
func (d *DBStore) Stop() {
	d.db.Close()
}

// Inserts a WakuMessage into the DB
func (d *DBStore) Put(cursor *pb.Index, pubsubTopic string, message *pb.WakuMessage) error {
	stmt, err := d.db.Prepare("INSERT INTO message (id, receiverTimestamp, senderTimestamp, contentTopic, pubsubTopic, payload, version) VALUES (?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(cursor.Digest, cursor.ReceiverTime, message.Timestamp, message.ContentTopic, pubsubTopic, message.Payload, message.Version)
	if err != nil {
		return err
	}

	return nil
}

// Returns all the stored WakuMessages
func (d *DBStore) GetAll() ([]persistence.StoredMessage, error) {
	rows, err := d.db.Query("SELECT id, receiverTimestamp, senderTimestamp, contentTopic, pubsubTopic, payload, version FROM message ORDER BY senderTimestamp ASC")
	if err != nil {
		return nil, err
	}

	var result []persistence.StoredMessage

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
			d.log.Fatal(err)
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
