package types

import (
	"github.com/libp2p/go-libp2p/core/peer"
)

var (
	EthereumWalletAddressLength = 42
	PeerIdLength                = 46
)

type WalletAddr string
type PeerId peer.ID

func (walletAddr WalletAddr) String() string {
	return string(walletAddr)
}

func (walletAddr WalletAddr) IsValid() bool {
	return EthereumWalletAddressLength == len(walletAddr.String())
}

func (peerId PeerId) String() string {
	return string(peerId)
}

func (peerId PeerId) Raw() peer.ID {
	return peer.ID(peerId)
}

func (peerId PeerId) IsValid() bool {
	return PeerIdLength == len(peerId.String())
}
