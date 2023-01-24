package crdt2

import (
	"fmt"

	mh "github.com/multiformats/go-multihash"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type cid mh.Multihash

func zapCid(key string, c mh.Multihash) zapcore.Field {
	return zap.Stringer(key, cid(c))
}

func (c cid) String() string {
	return shortenedCid(mh.Multihash(c))
}

type cidSlice []mh.Multihash

func (cids cidSlice) MarshalLogArray(enc zapcore.ArrayEncoder) error {
	for _, cid := range cids {
		enc.AppendString(shortenedCid(cid))
	}
	return nil
}

func zapCids(key string, cids ...mh.Multihash) zapcore.Field {
	return zap.Array(key, cidSlice(cids))
}

func shortenedCid(cid mh.Multihash) string {
	return fmt.Sprintf("%X…%X", []byte(cid[2:6]), []byte(cid[len(cid)-4:]))
}
