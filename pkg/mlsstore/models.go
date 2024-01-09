package mlsstore

import "github.com/uptrace/bun"

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

type Message struct {
	bun.BaseModel `bun:"table:messages"`

	TID       string `bun:",pk,notnull"`
	Topic     string `bun:"topic,notnull"`
	CreatedAt int64  `bun:"created_at,notnull"`
	Content   []byte `bun:"content,notnull,type:bytea"`
}
