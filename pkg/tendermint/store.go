package tendermint

import (
	"encoding/base64"
	"fmt"

	"github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"
	messagev1 "github.com/xmtp/proto/go/message_api/v1"
)

func buildKeyPrefix(topic string) string {
	encodedTopic := base64.StdEncoding.EncodeToString([]byte(topic))
	key := fmt.Sprintf("msg[%s,", encodedTopic)
	return key
}

func buildKey(env *messagev1.Envelope) (string, error) {
	encodedTopic := base64.StdEncoding.EncodeToString([]byte(env.ContentTopic))
	var msgCID string
	if len(env.Message) > 0 {
		id, err := buildCID(env.Message)
		if err != nil {
			return "", err
		}
		msgCID = id.String()
	}
	key := fmt.Sprintf("msg[%s,%d,%s]", encodedTopic, env.TimestampNs, msgCID)
	return key, nil
}

func buildCID(val []byte) (cid.Cid, error) {
	pref := cid.Prefix{
		Version:  1,
		Codec:    cid.Raw,
		MhType:   mh.SHA2_256,
		MhLength: -1, // default length
	}
	return pref.Sum(val)
}
