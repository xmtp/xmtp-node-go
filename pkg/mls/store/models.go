package store

import (
	"time"

	"github.com/uptrace/bun"
)

type InboxLogEntry struct {
	bun.BaseModel `bun:"table:inbox_log"`

	SequenceId          uint64
	InboxId             string
	ServerTimestampNs   int64
	IdentityUpdateProto []byte
}

type Installation struct {
	bun.BaseModel `bun:"table:installations"`

	ID                 []byte `bun:",pk,type:bytea"`
	WalletAddress      string `bun:"wallet_address,notnull"`
	CreatedAt          int64  `bun:"created_at,notnull"`
	UpdatedAt          int64  `bun:"updated_at,notnull"`
	RevokedAt          *int64 `bun:"revoked_at"`
	CredentialIdentity []byte `bun:"credential_identity,notnull,type:bytea"`

	KeyPackage []byte `bun:"key_package,notnull,type:bytea"`
	Expiration uint64 `bun:"expiration,notnull"`
}

type GroupMessage struct {
	bun.BaseModel `bun:"table:group_messages"`

	Id        uint64    `bun:",pk,notnull"`
	CreatedAt time.Time `bun:",notnull"`
	GroupId   []byte    `bun:",notnull,type:bytea"`
	Data      []byte    `bun:",notnull,type:bytea"`
}

type WelcomeMessage struct {
	bun.BaseModel `bun:"table:welcome_messages"`

	Id              uint64    `bun:",pk,notnull"`
	CreatedAt       time.Time `bun:",notnull"`
	InstallationKey []byte    `bun:",notnull,type:bytea"`
	Data            []byte    `bun:",notnull,type:bytea"`
	HpkePublicKey   []byte    `bun:"hpke_public_key,notnull,type:bytea"`
}
