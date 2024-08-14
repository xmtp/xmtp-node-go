// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0

package queries

import (
	"database/sql"
	"time"
)

type AddressLog struct {
	Address               string
	InboxID               []byte
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
	InboxID             []byte
	ServerTimestampNs   int64
	IdentityUpdateProto []byte
}

type Installation struct {
	ID         []byte
	CreatedAt  int64
	UpdatedAt  int64
	KeyPackage []byte
}

type WelcomeMessage struct {
	ID                      int64
	CreatedAt               time.Time
	InstallationKey         []byte
	Data                    []byte
	HpkePublicKey           []byte
	InstallationKeyDataHash []byte
}
