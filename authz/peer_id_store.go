package authz

import (
	"sync"

	"go.uber.org/zap"
)

type PeerWallet struct {
	WalletAddress string
}

type PeerIdStore interface {
	Get(peerId string) *PeerWallet
	Set(peerId, walletAddress string, allowed bool)
	Delete(peerId string)
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

func (s *MemoryPeerIdStore) Set(peerId, walletAddress string) {
	s.mutex.Lock()
	s.peers[peerId] = PeerWallet{
		WalletAddress: walletAddress,
	}
	s.log.Debug("stored peer",
		zap.String("peer_id", peerId),
		zap.String("wallet_address", walletAddress),
	)
	s.mutex.Unlock()
}

func (s *MemoryPeerIdStore) Delete(peerId string) {
	s.mutex.Lock()
	delete(s.peers, peerId)
	s.mutex.Unlock()
}
