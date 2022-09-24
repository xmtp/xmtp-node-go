package tendermint

import (
	"bytes"
	"fmt"

	"github.com/dgraph-io/badger"
	abci "github.com/tendermint/tendermint/abci/types"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	messagev1 "github.com/xmtp/proto/go/message_api/v1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type Application struct {
	log          *zap.Logger
	db           *badger.DB
	currentBatch *badger.Txn
}

var _ abcitypes.Application = (*Application)(nil)

func newApplication(log *zap.Logger, db *badger.DB) *Application {
	return &Application{
		log: log,
		db:  db,
	}
}

func (Application) Info(req abcitypes.RequestInfo) abcitypes.ResponseInfo {
	return abcitypes.ResponseInfo{}
}

func (Application) SetOption(req abcitypes.RequestSetOption) abcitypes.ResponseSetOption {
	return abcitypes.ResponseSetOption{}
}

func (app *Application) DeliverTx(req abcitypes.RequestDeliverTx) abcitypes.ResponseDeliverTx {
	app.log.Info("delivertx")

	code, log := app.isValid(req.Tx)
	if code != 0 {
		return abcitypes.ResponseDeliverTx{
			Code: code,
			Log:  log,
		}
	}

	// Parse envelope from the transaction.
	var env messagev1.Envelope
	err := proto.Unmarshal(req.Tx, &env)
	if err != nil {
		app.log.Error("unmarshaling envelope from transaction", zap.Error(err))
		return abcitypes.ResponseDeliverTx{
			Code: 1,
			Log:  "error parsing transaction",
		}
	}

	key, err := buildKey(&env)
	if err != nil {
		app.log.Error("creating message cid", zap.Error(err))
		return abcitypes.ResponseDeliverTx{
			Code: 1,
			Log:  "error creating cid",
		}
	}
	value := req.Tx
	app.log.Info("inserting envelope", zap.String("key", key), zap.String("value", string(value)))

	// Set stored message.
	err = app.currentBatch.Set([]byte(key), value)
	if err != nil {
		panic(err)
		// app.log.Error("setting db entry", zap.Error(err), zap.String("key", string(key)))
		// return abcitypes.ResponseDeliverTx{Code: 1}
	}

	events := []abci.Event{
		{
			Type: "envelope",
			Attributes: []abci.EventAttribute{
				{Key: []byte("topic"), Value: []byte(env.ContentTopic)},
				{Key: []byte("timestamp_ns"), Value: []byte(fmt.Sprintf("%d", env.TimestampNs))},
				{Key: []byte("message"), Value: env.Message},
			},
		},
	}

	return abcitypes.ResponseDeliverTx{
		Code:   0,
		Events: events,
	}
}

func (app *Application) isValid(tx []byte) (code uint32, log string) {
	var env messagev1.Envelope
	err := proto.Unmarshal(tx, &env)
	if err != nil {
		app.log.Info("invalid transaction", zap.Error(err))
		return 1, "error parsing transaction"
	}

	key, err := buildKey(&env)
	if err != nil {
		app.log.Error("creating message cid", zap.Error(err))
		return 1, "error creating cid"
	}
	value := tx

	// check if the same key=value already exists
	err = app.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil && err != badger.ErrKeyNotFound {
			return err
		}
		if err == nil {
			return item.Value(func(val []byte) error {
				if bytes.Equal(val, value) {
					code = 2
					log = "message already exists"
				}
				return nil
			})
		}
		return nil
	})
	if err != nil {
		app.log.Error("checking db", zap.Error(err))
		return 1, "error checking db"
	}

	return code, log
}

func (app *Application) CheckTx(req abcitypes.RequestCheckTx) abcitypes.ResponseCheckTx {
	app.log.Info("checktx")
	code, log := app.isValid(req.Tx)
	return abcitypes.ResponseCheckTx{
		Code:      code,
		Log:       log,
		GasWanted: 1,
	}
}

func (app *Application) Commit() abcitypes.ResponseCommit {
	app.currentBatch.Commit()
	return abcitypes.ResponseCommit{Data: []byte{}}
}

func (app *Application) Query(reqQuery abcitypes.RequestQuery) (resQuery abcitypes.ResponseQuery) {
	resQuery.Key = reqQuery.Data
	err := app.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(reqQuery.Data)
		if err != nil && err != badger.ErrKeyNotFound {
			return err
		}
		if err == badger.ErrKeyNotFound {
			resQuery.Log = "does not exist"
		} else {
			return item.Value(func(val []byte) error {
				resQuery.Log = "exists"
				resQuery.Value = val
				return nil
			})
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	return
}

func (Application) InitChain(req abcitypes.RequestInitChain) abcitypes.ResponseInitChain {
	return abcitypes.ResponseInitChain{}
}

func (app *Application) BeginBlock(req abcitypes.RequestBeginBlock) abcitypes.ResponseBeginBlock {
	app.currentBatch = app.db.NewTransaction(true)
	return abcitypes.ResponseBeginBlock{}
}

func (Application) EndBlock(req abcitypes.RequestEndBlock) abcitypes.ResponseEndBlock {
	return abcitypes.ResponseEndBlock{}
}

func (Application) ListSnapshots(abcitypes.RequestListSnapshots) abcitypes.ResponseListSnapshots {
	return abcitypes.ResponseListSnapshots{}
}

func (Application) OfferSnapshot(abcitypes.RequestOfferSnapshot) abcitypes.ResponseOfferSnapshot {
	return abcitypes.ResponseOfferSnapshot{}
}

func (Application) LoadSnapshotChunk(abcitypes.RequestLoadSnapshotChunk) abcitypes.ResponseLoadSnapshotChunk {
	return abcitypes.ResponseLoadSnapshotChunk{}
}

func (Application) ApplySnapshotChunk(abcitypes.RequestApplySnapshotChunk) abcitypes.ResponseApplySnapshotChunk {
	return abcitypes.ResponseApplySnapshotChunk{}
}
