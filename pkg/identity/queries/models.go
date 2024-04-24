// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0

package queries

import (
	"database/sql"
	"time"
)

type AddressLog struct {
	Address               string
	InboxID               string
	AssociationSequenceID sql.NullInt64
	RevocationSequenceID  sql.NullInt64
}

type GroupMessage struct {
	ID              int64
	CreatedAt       time.Time
	GroupID         []byte
	Data            []byte
	GroupIDDataHash []byte
}

type InboxLog struct {
	SequenceID          int64
	InboxID             string
	ServerTimestampNs   int64
	IdentityUpdateProto []byte
}

type Installation struct {
	ID                 []byte
	WalletAddress      string
	CreatedAt          int64
	UpdatedAt          int64
	CredentialIdentity []byte
	RevokedAt          sql.NullInt64
	KeyPackage         []byte
	Expiration         int64
}

type WelcomeMessage struct {
	ID                      int64
	CreatedAt               time.Time
	InstallationKey         []byte
	Data                    []byte
	InstallationKeyDataHash []byte
	HpkePublicKey           []byte
}