package crdt

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	ipfslite "github.com/hsanjuan/ipfs-lite"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	badger "github.com/ipfs/go-ds-badger"
	crdt "github.com/ipfs/go-ds-crdt"
	"github.com/libp2p/go-libp2p-kad-dht/dual"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	messagev1 "github.com/xmtp/proto/v3/go/message_api/v1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

const (
	DefaultPubsubTopic = "/xmtp/1/default-xmtp/proto"
)

type Node struct {
	ctx         context.Context
	log         *zap.Logger
	store       *badger.Datastore
	host        host.Host
	dht         *dual.DHT
	broadcaster *crdt.PubSubBroadcaster
	ipfs        *ipfslite.Peer

	crdtByTopic   map[string]*crdt.Datastore
	crdtByTopicMu sync.Mutex

	EnvC chan *messagev1.Envelope
}

type Options struct {
	NodeKey            string
	DataPath           string
	P2PPort            int
	P2PPersistentPeers []string
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
	nodeId, err := peer.IDFromPublicKey(privKey.GetPublic())
	if err != nil {
		return nil, errors.Wrap(err, "getting peer id from key")
	}
	log.Info("starting", zap.String("node_id", nodeId.Pretty()))

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
	log = log.With(zap.String("node", host.ID().Pretty()))
	log.Info("listening", zap.Strings("addresses", listenAddresses(host)))
	ipfs, err := ipfslite.New(ctx, store, nil, host, dht, nil)
	if err != nil {
		return nil, errors.Wrap(err, "initializing ipfslite")
	}

	// Log connected peers periodically.
	go func() {
		for {
			peers := []*peer.AddrInfo{}
			peerAddrs := []string{}
			for _, conn := range host.Network().Conns() {
				peers = append(peers, &peer.AddrInfo{
					ID:    conn.RemotePeer(),
					Addrs: []multiaddr.Multiaddr{conn.RemoteMultiaddr()},
				})
				peerAddrs = append(peerAddrs, fmt.Sprintf("%s/%s", conn.RemoteMultiaddr().String(), conn.RemotePeer().Pretty()))
			}
			log.Info("total connected peers", zap.Int("total_peers", len(peers)), zap.Strings("peers", peerAddrs))
			time.Sleep(60 * time.Second)
		}
	}()

	// Configure bootstrap peers.
	persistentPeers := make([]peer.AddrInfo, 0, len(options.P2PPersistentPeers))
	for _, addr := range options.P2PPersistentPeers {
		maddr, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			return nil, errors.Wrap(err, "parsing persistent peer address")
		}
		peer, err := peer.AddrInfoFromP2pAddr(maddr)
		if err != nil {
			return nil, errors.Wrap(err, "getting persistent peer address info")
		}
		if peer == nil {
			return nil, fmt.Errorf("persistent peer address info is nil: %s", addr)
		}
		if peer.ID == nodeId {
			continue
		}
		persistentPeers = append(persistentPeers, *peer)
		host.ConnManager().TagPeer(peer.ID, "keep", 100)
	}
	if len(persistentPeers) > 0 {
		ipfs.Bootstrap(persistentPeers)
	}

	// Add a ping service and ping persistent peers regularly, reconnecting if
	// necessary.
	pinger := ping.NewPingService(host)
	for _, pp := range persistentPeers {
		go func(pp peer.AddrInfo) {
			log := log.With(zap.String("peer_id", pp.ID.Pretty()))
			pingCtx, cancel := context.WithCancel(ctx)
			pingC := pinger.Ping(pingCtx, pp.ID)
			for {
				connected := map[peer.ID]bool{}
				for _, conn := range host.Network().Conns() {
					connected[conn.RemotePeer()] = true
				}
				if !connected[pp.ID] {
					log.Info("reconnecting to peer")
					err = host.Connect(ctx, pp)
					if err != nil {
						log.Warn("error connecting to peer", zap.Error(err))
					}
					cancel()
					pingCtx, cancel = context.WithCancel(ctx)
					pingC = pinger.Ping(pingCtx, pp.ID)
					time.Sleep(5 * time.Second)
					continue
				}

				log.Debug("pinging peer")
				select {
				case <-ctx.Done():
					cancel()
					return
				case res := <-pingC:
					if res.Error != nil {
						log.Warn("peer ping error", zap.Error(res.Error))
						if strings.Contains(res.Error.Error(), "resource scope closed") {
							err := host.Network().ClosePeer(pp.ID)
							if err != nil {
								log.Info("error closing peer", zap.Error(err))
							}
							cancel()
							pingCtx, cancel = context.WithCancel(ctx)
							pingC = pinger.Ping(pingCtx, pp.ID)
						}
					} else {
						log.Debug("peer ping ok", zap.Duration("rtt", res.RTT))
					}
					time.Sleep(5 * time.Second)
				case <-time.After(time.Second * 5):
					log.Warn("peer ping timeout")
				}
			}
		}(pp)
	}

	// Initialize pubsub broadcaster.
	psub, err := pubsub.NewGossipSub(ctx, host)
	// psub, err := pubsub.NewFloodSub(ctx, host)
	if err != nil {
		return nil, errors.Wrap(err, "initializing libp2p gossipsub")
	}
	pubsubBC, err := crdt.NewPubSubBroadcaster(ctx, psub, DefaultPubsubTopic)
	if err != nil {
		return nil, errors.Wrap(err, "initializing crdt pubsub broadcaster")
	}

	n := &Node{
		ctx:         ctx,
		log:         log,
		store:       store,
		host:        host,
		dht:         dht,
		broadcaster: pubsubBC,
		ipfs:        ipfs,
		crdtByTopic: map[string]*crdt.Datastore{},

		EnvC: make(chan *messagev1.Envelope),
	}

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
	crdt, err := n.getOrCreate(env.ContentTopic)
	if err != nil {
		return err
	}
	envBytes, err := proto.Marshal(env)
	if err != nil {
		return err
	}
	storeKey, err := buildMessageStoreKey(env)
	if err != nil {
		return errors.Wrap(err, "building message store key")
	}
	err = crdt.Put(ctx, storeKey, envBytes)
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
	orders := []query.Order{}
	if req.PagingInfo != nil && req.PagingInfo.Direction == messagev1.SortDirection_SORT_DIRECTION_DESCENDING {
		orders = append(orders, query.OrderByKeyDescending{})
	}
	crdt, err := n.getOrCreate(topic)
	if err != nil {
		return nil, nil, err
	}
	res, err := crdt.Query(ctx, query.Query{
		Prefix: buildMessageQueryPrefix(topic),
		Orders: orders,
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
		priv, err := GenerateNodeKey()
		if err != nil {
			return nil, err
		}

		return priv, nil
	}

	keyBytes, err := hex.DecodeString(key)
	if err != nil {
		return nil, errors.Wrap(err, "decoding node key")
	}
	return crypto.UnmarshalPrivateKey(keyBytes)
}

func listenAddresses(host host.Host) []string {
	exclude := map[string]bool{
		"/p2p-circuit": true,
	}
	addrs := []string{}
	for _, ma := range host.Network().ListenAddresses() {
		addr := ma.String()
		if exclude[addr] {
			continue
		}
		addrs = append(addrs, addr)
	}
	return addrs
}

func (n *Node) getOrCreate(topic string) (*crdt.Datastore, error) {
	n.crdtByTopicMu.Lock()
	defer n.crdtByTopicMu.Unlock()

	if _, exists := n.crdtByTopic[topic]; !exists {
		opts := crdt.DefaultOptions()
		opts.Logger = &logger{n.log}
		opts.RebroadcastInterval = 5 * time.Second
		opts.MultiHeadProcessing = true

		opts.PutHook = func(key datastore.Key, value []byte) {
			var env messagev1.Envelope
			err := proto.Unmarshal(value, &env)
			if err != nil {
				n.log.Info("parsing envelope", zap.Error(err))
			} else {
				n.EnvC <- &env
			}
		}
		crdt, err := crdt.New(n.store, datastore.NewKey("topic/"+topic), n.ipfs, n.broadcaster, opts)
		if err != nil {
			return nil, errors.Wrap(err, "initializing crdt")
		}
		n.crdtByTopic[topic] = crdt
	}

	return n.crdtByTopic[topic], nil
}
