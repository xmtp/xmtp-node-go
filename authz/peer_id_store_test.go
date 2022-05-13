package authz

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const (
	WALLET_ADDRESS = "0x1234"
	PEER_ID        = "P5678"
)

func TestGetSet(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	store := NewMemoryPeerIdStore(logger)

	store.Set(PEER_ID, WALLET_ADDRESS)

	entry := store.Get(PEER_ID)
	require.Equal(t, entry.WalletAddress, WALLET_ADDRESS)
	// Test overwriting the existing entry
	store.Set(PEER_ID, "0xfoo")
	entry = store.Get(PEER_ID)
	require.Equal(t, entry.WalletAddress, "0xfoo")
}

func TestNil(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	store := NewMemoryPeerIdStore(logger)

	entry := store.Get(PEER_ID)
	require.Nil(t, entry)
}

func TestConcurrent(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	store := NewMemoryPeerIdStore(logger)
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			store.Set(fmt.Sprintf("peer-%d", idx), WALLET_ADDRESS)
		}(i)
	}
	wg.Wait()
	for i := 0; i < 100; i++ {
		val := store.Get(fmt.Sprintf("peer-%d", i))
		require.NotNil(t, val)
		require.Equal(t, val.WalletAddress, WALLET_ADDRESS)
	}
}
