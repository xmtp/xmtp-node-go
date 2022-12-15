package crdt

import (
	"context"
	"fmt"
	"testing"

	badger "github.com/ipfs/go-ds-badger3"
	"github.com/stretchr/testify/require"
	v1 "github.com/xmtp/proto/go/message_api/v1"
	"google.golang.org/protobuf/proto"
)

func Test_Query(t *testing.T) {
	db, data := newInMemoryDatastore(t, 10, 5)
	defer db.Close()
	t.Run("read whole topic", func(t *testing.T) {
		requireQueryResult(t, db,
			&v1.QueryRequest{
				ContentTopics: []string{"m-0003"},
			},
			data["m-0003"],
		)
	})
	t.Run("read whole topic descending", func(t *testing.T) {
		requireQueryResult(t, db,
			&v1.QueryRequest{
				ContentTopics: []string{"m-0003"},
				PagingInfo: &v1.PagingInfo{
					Direction: v1.SortDirection_SORT_DIRECTION_DESCENDING,
				},
			},
			reversed(data["m-0003"]),
		)
	})
	t.Run("read first 2 from topic", func(t *testing.T) {
		requireQueryResult(t, db,
			&v1.QueryRequest{
				ContentTopics: []string{"m-0003"},
				PagingInfo: &v1.PagingInfo{
					Limit: 2,
				},
			},
			data["m-0003"][:2],
		)
	})
}

func reversed(in []*v1.Envelope) []*v1.Envelope {
	reversed := make([]*v1.Envelope, len(in))
	for i, v := range in {
		reversed[len(in)-1-i] = v
	}
	return reversed
}

func requireQueryResult(t *testing.T, db *badger.Datastore, q *v1.QueryRequest, envs []*v1.Envelope) {
	ctx := context.Background()
	qrs, err := db.Query(ctx, buildMessageQuery(q))
	require.NoError(t, err)
	defer qrs.Close()
	i := 0
	for res := range qrs.Next() {
		require.NoError(t, res.Error)
		require.Less(t, i, len(envs))
		var env v1.Envelope
		require.NoError(t, proto.Unmarshal(res.Value, &env))
		exp := envs[i]
		require.Equal(t, exp.ContentTopic, env.ContentTopic)
		require.Equal(t, exp.TimestampNs, env.TimestampNs)
		require.Equal(t, exp.Message, env.Message)
		i++
	}
}

// sample db with topicsN topics and up to envelopesN messages in each
func newInMemoryDatastore(t *testing.T, topicsN, envelopesN int) (db *badger.Datastore, data map[string][]*v1.Envelope) {
	opts := badger.DefaultOptions
	opts.Options = opts.Options.WithInMemory(true)
	db, err := badger.NewDatastore("", &opts)
	require.NoError(t, err)
	data = make(map[string][]*v1.Envelope)
	if testing.Verbose() {
		t.Log("Populating store")
	}
	for i := 0; i < topicsN; i++ {
		topic := fmt.Sprintf("m-%04d", i)
		if testing.Verbose() {
			t.Logf("Topic: %s", topic)
		}
		for j := 0; j <= i%envelopesN; j++ {
			message := fmt.Sprintf("%04d/%010d", i, j)
			env := &v1.Envelope{
				ContentTopic: topic,
				TimestampNs:  uint64(j*100 + i*10),
				Message:      []byte(message),
			}
			eb, err := proto.Marshal(env)
			require.NoError(t, err)
			key, err := buildMessageStoreKey(env)
			require.NoError(t, err)
			err = db.Put(context.Background(), key, eb)
			require.NoError(t, err)
			if testing.Verbose() {
				t.Logf("\t%s\n\t\t%5d %s", key, env.TimestampNs, message)
			}
			data[topic] = append(data[topic], env)
		}
	}
	if testing.Verbose() {
		t.Log("Done")
	}
	return db, data
}
