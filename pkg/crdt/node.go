package crdt

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	ipfslite "github.com/hsanjuan/ipfs-lite"
	"github.com/ipfs/go-datastore"
	badger "github.com/ipfs/go-ds-badger3"
	crdt "github.com/ipfs/go-ds-crdt"
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-kad-dht/dual"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
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
	ctx   context.Context
	log   *zap.Logger
	store *badger.Datastore
	crdt  *crdt.Datastore
	host  host.Host
	dht   *dual.DHT

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
	log.Info("listening", zap.Strings("addresses", listenAddresses(host)))
	ipfs, err := ipfslite.New(ctx, store, nil, host, dht, nil)
	if err != nil {
		return nil, errors.Wrap(err, "initializing ipfslite")
	}

	// Log connected peers periodically.
	go func() {
		for {
			peersByID := map[string]int{}
			peers := []*peer.AddrInfo{}
			peerAddrs := []string{}
			for _, conn := range host.Network().Conns() {
				peerID := conn.RemotePeer()
				if peersByID[peerID.Pretty()] > 0 {
					log.Warn("Duplicate peer connection found, disconnecting peer", zap.String("node_id", peerID.Pretty()))
					err := host.Network().ClosePeer(peerID)
					if err != nil {
						log.Info("closing peer", zap.Error(err), zap.String("node_id", peerID.Pretty()))
					}
					break
				}
				peersByID[peerID.Pretty()]++
				peers = append(peers, &peer.AddrInfo{
					ID:    conn.RemotePeer(),
					Addrs: []multiaddr.Multiaddr{conn.RemoteMultiaddr()},
				})
				peerAddrs = append(peerAddrs, fmt.Sprintf("%s/%s", conn.RemoteMultiaddr().String(), conn.RemotePeer().Pretty()))
			}
			log.Info("total connected peers", zap.Int("total_peers", len(peers)), zap.Strings("peers", peerAddrs))
			time.Sleep(10 * time.Second)
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

	go n.persistentPeersConnectLoop(persistentPeers)

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
	res, err := n.crdt.Query(ctx, buildMessageQuery(req))
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

func (n *Node) persistentPeersConnectLoop(peers []peer.AddrInfo) {
	for _, peer := range peers {
		err := n.host.Connect(n.ctx, peer)
		if err != nil {
			n.log.Error("dialing persistent peer", zap.Error(err), zap.String("peer_addr", peer.String()))
		}
	}

	ticker := time.NewTicker(2 * time.Second)
	for {
		select {
		case <-n.ctx.Done():
			return
		case <-ticker.C:
			peerIDs := map[peer.ID]struct{}{}
			for _, peerID := range n.host.Network().Peers() {
				peerIDs[peerID] = struct{}{}
			}
			for _, peer := range peers {
				if _, exists := peerIDs[peer.ID]; exists {
					continue
				}
				err := n.host.Connect(n.ctx, peer)
				if err != nil {
					n.log.Error("dialing persistent peer", zap.Error(err), zap.String("peer_addr", peer.String()))
				}
			}
		}
	}
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
