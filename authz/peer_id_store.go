package authz

import (
	"sync"

	"go.uber.org/zap"
)

type PeerWallet struct {
	WalletAddress string
	IsAllowed     bool
}

type PeerIdStore interface {
	Get(peerId string) *PeerWallet
	Set(peerId, walletAddress string, allowed bool)
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

func (s *MemoryPeerIdStore) Get(peerId string) *PeerWallet {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if res, ok := s.peers[peerId]; ok {
		return &res
	}
	return nil
}

func (s *MemoryPeerIdStore) Set(peerId, walletAddress string, allowed bool) {
	s.mutex.Lock()
	s.peers[peerId] = PeerWallet{
		WalletAddress: walletAddress,
		IsAllowed:     allowed,
	}
	s.log.Debug("stored peer",
		zap.String("peer_id", peerId),
		zap.String("wallet_address", walletAddress),
		zap.Bool("allowed", allowed),
	)
	s.mutex.Unlock()
}
