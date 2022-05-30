package types

import "github.com/libp2p/go-libp2p-core/peer"

type WalletAddr string
type PeerId peer.ID

func (walletAddr WalletAddr) String() string {
	return string(walletAddr)
}

func InvalidWalletAddr() WalletAddr {
	return ""
}

func (peerId PeerId) String() string {
	return string(peerId)
}

func (peerId PeerId) Raw() peer.ID {
	return peer.ID(peerId)
}

func InvalidPeerId() PeerId {
	return ""
}
