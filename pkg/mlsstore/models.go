package mlsstore

import "github.com/uptrace/bun"

type Installation struct {
	bun.BaseModel `bun:"table:installations"`

	ID            string `bun:",pk"`
	WalletAddress string `bun:"wallet_address,notnull"`
	CreatedAt     int64  `bun:"created_at,notnull"`
	RevokedAt     *int64 `bun:"revoked_at"`
}

type KeyPackage struct {
	bun.BaseModel `bun:"table:key_packages"`

	ID             string `bun:",pk"` // ID is the hash of the data field
	InstallationId string `bun:"installation_id,notnull"`
	CreatedAt      int64  `bun:"created_at,notnull"`
	ConsumedAt     *int64 `bun:"consumed_at"`
	IsLastResort   bool   `bun:"is_last_resort,notnull"`
	Data           []byte `bun:"data,notnull,type:bytea"`
}
