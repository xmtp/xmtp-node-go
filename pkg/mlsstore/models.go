package mlsstore

import "github.com/uptrace/bun"

type Installation struct {
	bun.BaseModel `bun:"table:installations"`

	ID                 []byte `bun:",pk,type:bytea"`
	WalletAddress      string `bun:"wallet_address,notnull"`
	CreatedAt          int64  `bun:"created_at,notnull"`
	RevokedAt          *int64 `bun:"revoked_at"`
	CredentialIdentity []byte `bun:"credential_identity,notnull,type:bytea"`
}

type KeyPackage struct {
	bun.BaseModel `bun:"table:key_packages"`

	ID             string `bun:",pk"` // ID is the hash of the data field
	InstallationId []byte `bun:"installation_id,notnull,type:bytea"`
	CreatedAt      int64  `bun:"created_at,notnull"`
	ConsumedAt     *int64 `bun:"consumed_at"`
	NotConsumed    bool   `bun:"not_consumed,default:true"`
	IsLastResort   bool   `bun:"is_last_resort,notnull"`
	Data           []byte `bun:"data,notnull,type:bytea"`
}
