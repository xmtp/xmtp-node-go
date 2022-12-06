package crdt

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	ipfslite "github.com/hsanjuan/ipfs-lite"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	badger "github.com/ipfs/go-ds-badger"
	crdt "github.com/ipfs/go-ds-crdt"
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-kad-dht/dual"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	messagev1 "github.com/xmtp/proto/go/message_api/v1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

const (
	DefaultPubsubTopic = "/xmtp/1/default-xmtp/proto"
)

type Node struct {
	ctx   context.Context
	log   *zap.Logger
	store *badger.Datastore
	crdt  *crdt.Datastore
	host  host.Host
	dht   *dual.DHT

	EnvC chan *messagev1.Envelope
}

func NewNode(ctx context.Context, log *zap.Logger, options Options) (*Node, error) {
	log = log.Named("crdt")

	// Initialize embedded datastore.
	if options.DataPath == "" {
		return nil, errors.New("missing data-path argument")
	}
	err := os.MkdirAll(options.DataPath, os.ModePerm)
	if err != nil {
		return nil, err
	}
	store, err := badger.NewDatastore(options.DataPath, &badger.DefaultOptions)
	if err != nil {
		return nil, errors.Wrap(err, "initializing datastore")
	}

	// Get libp2p peer ID.
	privKey, err := getOrCreatePrivateKey(options.NodeKey)
	if err != nil {
		return nil, errors.Wrap(err, "creating libp2p crypto key")
	}
	pid, err := peer.IDFromPublicKey(privKey.GetPublic())
	if err != nil {
		return nil, errors.Wrap(err, "getting peer id from key")
	}
	log.Info("starting", zap.String("peer_id", pid.Pretty()))

	// Initialize IPFS-lite libp2p.
	listenAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", options.P2PPort))
	if err != nil {
		return nil, errors.Wrap(err, "initializing crdt listen addr")
	}
	host, dht, err := ipfslite.SetupLibp2p(
		ctx,
		privKey,
		nil,
		[]multiaddr.Multiaddr{listenAddr},
		nil,
		ipfslite.Libp2pOptionsExtra...,
	)
	if err != nil {
		return nil, errors.Wrap(err, "initializing ipfslite libp2p")
	}
	ipfs, err := ipfslite.New(ctx, store, nil, host, dht, nil)
	if err != nil {
		return nil, errors.Wrap(err, "initializing ipfslite")
	}

	// Log connected peers periodically.
	go func() {
		for {
			peers := []*peer.AddrInfo{}
			for _, c := range host.Network().Conns() {
				peers = append(peers, &peer.AddrInfo{
					ID:    c.RemotePeer(),
					Addrs: []multiaddr.Multiaddr{c.RemoteMultiaddr()},
				})
			}
			log.Info("total connected peers", zap.Int("total_peers", len(peers)))
			time.Sleep(10 * time.Second)
		}
	}()

	// Configure bootstrap peers.
	bootstrapPeers := make([]peer.AddrInfo, len(options.P2PBootstrapNodes))
	for i, addr := range options.P2PBootstrapNodes {
		maddr, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			return nil, errors.Wrap(err, "parsing bootstrap node address")
		}
		info, err := peer.AddrInfoFromP2pAddr(maddr)
		if err != nil {
			return nil, errors.Wrap(err, "getting bootstrap node address info")
		}
		if info == nil {
			return nil, fmt.Errorf("bootstrap node address info is nil: %s", addr)
		}
		bootstrapPeers[i] = *info
		host.ConnManager().TagPeer(info.ID, "keep", 100)
	}
	if len(bootstrapPeers) > 0 {
		ipfs.Bootstrap(bootstrapPeers)
	}

	// Initialize pubsub broadcaster.
	psub, err := pubsub.NewGossipSub(ctx, host)
	if err != nil {
		return nil, errors.Wrap(err, "initializing libp2p gossipsub")
	}
	pubsubBC, err := crdt.NewPubSubBroadcaster(ctx, psub, DefaultPubsubTopic)
	if err != nil {
		return nil, errors.Wrap(err, "initializing crdt pubsub broadcaster")
	}

	// Initialize crdt/dag server.
	opts := crdt.DefaultOptions()
	opts.Logger = logging.Logger("xmtpd")
	opts.RebroadcastInterval = 5 * time.Second

	envC := make(chan *messagev1.Envelope)
	opts.PutHook = func(key datastore.Key, value []byte) {
		var env messagev1.Envelope
		err := proto.Unmarshal(value, &env)
		if err != nil {
			log.Info("parsing envelope", zap.Error(err))
		} else {
			envC <- &env
		}
	}
	crdt, err := crdt.New(store, datastore.NewKey("crdt"), ipfs, pubsubBC, opts)
	if err != nil {
		return nil, errors.Wrap(err, "initializing crdt")
	}

	n := &Node{
		ctx:   ctx,
		log:   log,
		store: store,
		crdt:  crdt,
		host:  host,
		dht:   dht,
		EnvC:  envC,
	}

	go n.bootstrapNodesConnectLoop(options.P2PBootstrapNodes)

	return n, nil
}

func (n *Node) Close() error {
	// Close envs channel.
	if n.EnvC != nil {
		close(n.EnvC)
	}

	// Close host.
	if n.host != nil {
		err := n.host.Close()
		if err != nil {
			return errors.Wrap(err, "closing crdt host")
		}
	}

	// Close DHT.
	if n.dht != nil {
		err := n.dht.Close()
		if err != nil {
			return errors.Wrap(err, "closing crdt dht")
		}
	}

	// // Close crdt.
	// TODO: this causes Close to hang, unclear why
	// if n.crdt != nil {
	// 	err := n.crdt.Close()
	// 	if err != nil {
	// 		return errors.Wrap(err, "closing crdt crdt")
	// 	}
	// }

	// Close datastore.
	if n.store != nil {
		err := n.store.Close()
		if err != nil {
			return errors.Wrap(err, "closing crdt datastore")
		}
	}

	return nil
}

func (n *Node) Publish(ctx context.Context, env *messagev1.Envelope) error {
	envBytes, err := proto.Marshal(env)
	if err != nil {
		return err
	}
	storeKey, err := buildMessageStoreKey(env)
	if err != nil {
		return errors.Wrap(err, "building message store key")
	}
	err = n.crdt.Put(ctx, storeKey, envBytes)
	if err != nil {
		return err
	}
	return nil
}

func (n *Node) Query(ctx context.Context, req *messagev1.QueryRequest) ([]*messagev1.Envelope, *messagev1.PagingInfo, error) {
	var topic string
	if len(req.ContentTopics) > 0 {
		topic = req.ContentTopics[0] // TODO
	}
	// TODO: sorting, start/end time filtering
	res, err := n.crdt.Query(ctx, query.Query{
		Prefix: buildMessageQueryPrefix(topic),
	})
	if err != nil {
		return nil, nil, err
	}

	entries, err := res.Rest()
	if err != nil {
		return nil, nil, err
	}

	envs := []*messagev1.Envelope{}
	for _, entry := range entries {
		var env messagev1.Envelope
		err := proto.Unmarshal(entry.Value, &env)
		if err != nil {
			n.log.Info("parsing stored envelope", zap.Error(err))
		}
		envs = append(envs, &env)
	}

	// TODO: pagingInfo

	return envs, nil, nil
}

func getOrCreatePrivateKey(key string) (crypto.PrivKey, error) {
	if key == "" {
		priv, _, err := crypto.GenerateKeyPair(crypto.Ed25519, 1)
		if err != nil {
			return nil, err
		}

		// keyBytes, err := crypto.MarshalPrivateKey(priv)
		// if err != nil {
		// 	return nil, err
		// }
		// fmt.Println("NODE KEY:", hex.EncodeToString(keyBytes))

		return priv, nil
	}

	keyBytes, err := hex.DecodeString(key)
	if err != nil {
		return nil, errors.Wrap(err, "decoding node key")
	}
	return crypto.UnmarshalPrivateKey(keyBytes)
}

func (n *Node) bootstrapNodesConnectLoop(bootstrapNodes []string) {
	peers := make([]peer.AddrInfo, len(bootstrapNodes))

	var selfAddr string
	for i, peerAddr := range bootstrapNodes {
		maddr, err := multiaddr.NewMultiaddr(peerAddr)
		if err != nil {
			n.log.Error("parsing bootstrap address", zap.Error(err), zap.String("peer_addr", peerAddr))
		}
		info, err := peer.AddrInfoFromP2pAddr(maddr)
		if err != nil {
			n.log.Error("getting bootstrap node address info", zap.Error(err), zap.String("peer_addr", peerAddr))
		}
		if info == nil {
			n.log.Error("bootstrap node address info is nil", zap.Error(err), zap.String("peer_addr", peerAddr))
		} else {
			peers[i] = *info
			err := n.host.Connect(n.ctx, *info)
			if err != nil {
				if strings.Contains(err.Error(), "dial to self") {
					selfAddr = peerAddr
				}
				n.log.Error("dialing bootstrap node", zap.Error(err), zap.String("peer_addr", info.String()))
			}
		}
	}

	ticker := time.NewTicker(500 * time.Millisecond)
	for {
		select {
		case <-n.ctx.Done():
			return
		case <-ticker.C:
			peerIDs := map[peer.ID]struct{}{}
			for _, peerID := range n.host.Network().Peers() {
				peerIDs[peerID] = struct{}{}
			}
			for i, addr := range bootstrapNodes {
				if addr == selfAddr {
					continue
				}
				pi := peers[i]
				if _, exists := peerIDs[pi.ID]; exists {
					continue
				}
				err := n.host.Connect(n.ctx, pi)
				if err != nil {
					n.log.Error("dialing bootstrap node", zap.Error(err), zap.String("peer_addr", pi.String()))
				}
			}
		}
	}
}
