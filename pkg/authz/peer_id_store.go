package authz

import (
	"context"
	"sync"
	"time"

	"github.com/xmtp/xmtp-node-go/pkg/tracing"
	"go.uber.org/zap"
)

const (
	RETENTION_PERIOD = 24 * time.Hour
	PURGE_INTERVAL   = 10 * time.Minute
)

type PeerWallet struct {
	WalletAddress string
	createdAt     time.Time
}

type PeerIdStore interface {
	Get(peerId string) *PeerWallet
	Set(peerId, walletAddress string, allowed bool)
	Start(ctx context.Context)
}

type MemoryPeerIdStore struct {
	peers map[string]PeerWallet
	mutex sync.RWMutex
	log   *zap.Logger
}

func NewMemoryPeerIdStore(log *zap.Logger) *MemoryPeerIdStore {
	store := new(MemoryPeerIdStore)
	store.log = log.Named("peer-id-store")
	store.peers = make(map[string]PeerWallet)

	return store
}

func (s *MemoryPeerIdStore) Start(ctx context.Context) {
	go tracing.PanicsDo(ctx, "authz-purge-loop", func(ctx context.Context) { s.purgeLoop(ctx) })
}

func (s *MemoryPeerIdStore) Get(peerId string) *PeerWallet {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if res, ok := s.peers[peerId]; ok {
		return &res
	}
	return nil
}

func (s *MemoryPeerIdStore) Set(peerId, walletAddress string) {
	s.mutex.Lock()
	s.peers[peerId] = PeerWallet{
		WalletAddress: walletAddress,
		createdAt:     time.Now(),
	}
	s.log.Debug("stored peer",
		zap.String("peer_id", peerId),
		zap.String("wallet_address", walletAddress),
	)
	s.mutex.Unlock()
}

func (s *MemoryPeerIdStore) delete(peerId string) {
	s.mutex.Lock()
	delete(s.peers, peerId)
	s.mutex.Unlock()
}

func (s *MemoryPeerIdStore) purgeExpired() {
	s.mutex.RLock()
	cutOff := time.Now().Add(RETENTION_PERIOD * -1)
	numPurged := 0
	for peerId, peerWallet := range s.peers {
		if peerWallet.createdAt.Before(cutOff) {
			numPurged += 1
			// Unlock the read mutex so we can obtain a lock in the write mutex
			s.mutex.RUnlock()
			s.delete(peerId)
			s.mutex.RLock()
		}
	}

	s.log.Info("record purge complete", zap.Int("num_records_purged", numPurged))
	s.mutex.RUnlock()
}

func (s *MemoryPeerIdStore) purgeLoop(ctx context.Context) {
	ticker := time.NewTicker(PURGE_INTERVAL)
	for {
		select {
		case <-ctx.Done():
			s.log.Info("Stopping purge loop")
		case <-ticker.C:
			s.purgeExpired()
		}
	}
}
