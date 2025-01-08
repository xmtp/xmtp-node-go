package main

import (
	"context"
	"log"
	"time"

	identityProto "github.com/xmtp/xmtp-node-go/pkg/proto/identity/api/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	ctx := context.Background()

	// Connect to the identity service
	// To connect to the dev or production environment, you will need to use secure credentials
	conn, err := grpc.Dial("localhost:5556", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Create identity service client
	client := identityProto.NewIdentityApiClient(conn)

	// Subscribe to association changes
	stream, err := client.SubscribeAssociationChanges(ctx, &identityProto.SubscribeAssociationChangesRequest{})
	if err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}

	// Read from the stream
	for {
		msg, err := stream.Recv()
		if err != nil {
			log.Fatalf("Failed to receive message: %v", err)
		}
		dateString := nsToDate(msg.TimestampNs)

		if msg.GetAccountAddressAssociation() != nil {
			accountAddressAssociation := msg.GetAccountAddressAssociation()
			log.Printf("Wallet %s was associated with inbox %s at %s",
				accountAddressAssociation.AccountAddress,
				accountAddressAssociation.InboxId,
				dateString,
			)
		}
		if msg.GetAccountAddressRevocation() != nil {
			accountAddressRevocation := msg.GetAccountAddressRevocation()
			log.Printf("Wallet %s was revoked from inbox %s at %s",
				accountAddressRevocation.AccountAddress,
				accountAddressRevocation.InboxId,
				dateString,
			)
		}
	}
}

func nsToDate(ns uint64) string {
	return time.Unix(0, int64(ns)).Format(time.RFC3339)
}
