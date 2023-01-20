package crdt2

import (
	mh "github.com/multiformats/go-multihash"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type cidSlice []mh.Multihash

func (cids cidSlice) MarshalLogArray(enc zapcore.ArrayEncoder) error {
	for _, cid := range cids {
		enc.AppendString(cid.String())
	}
	return nil
}

func zapCids(cids ...mh.Multihash) zapcore.Field {
	return zap.Array("cids", cidSlice(cids))
}
