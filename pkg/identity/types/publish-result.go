package types

import (
	identity "github.com/xmtp/xmtp-node-go/pkg/proto/identity/api/v1"
)

type PublishIdentityUpdateResult struct {
	InboxID string

	NewAddresses         []string
	RevokedAddresses     []string
	NewInstallations     [][]byte
	RevokedInstallations [][]byte
	TimestampNs          uint64
}

func NewPublishIdentityUpdateResult(inboxID string, timestampNs uint64, newAddresses []string, revokedAddresses []string, newInstallations [][]byte, revokedInstallations [][]byte) *PublishIdentityUpdateResult {
	return &PublishIdentityUpdateResult{
		InboxID:              inboxID,
		TimestampNs:          timestampNs,
		NewAddresses:         newAddresses,
		RevokedAddresses:     revokedAddresses,
		NewInstallations:     newInstallations,
		RevokedInstallations: revokedInstallations,
	}
}

func (p *PublishIdentityUpdateResult) GetChanges() []*identity.SubscribeAssociationChangesResponse {
	out := make([]*identity.SubscribeAssociationChangesResponse, 0)

	for _, newAddress := range p.NewAddresses {
		out = append(out, &identity.SubscribeAssociationChangesResponse{
			TimestampNs: p.TimestampNs,
			Change: &identity.SubscribeAssociationChangesResponse_AccountAddressAssociation_{
				AccountAddressAssociation: &identity.SubscribeAssociationChangesResponse_AccountAddressAssociation{
					InboxId:        p.InboxID,
					AccountAddress: newAddress,
				},
			},
		})
	}

	for _, revokedAddress := range p.RevokedAddresses {
		out = append(out, &identity.SubscribeAssociationChangesResponse{
			TimestampNs: p.TimestampNs,
			Change: &identity.SubscribeAssociationChangesResponse_AccountAddressRevocation_{
				AccountAddressRevocation: &identity.SubscribeAssociationChangesResponse_AccountAddressRevocation{
					InboxId:        p.InboxID,
					AccountAddress: revokedAddress,
				},
			},
		})
	}

	return out
}
