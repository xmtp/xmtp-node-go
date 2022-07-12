package authz

import (
	"time"

	"github.com/uptrace/bun"
)

type WalletAddress struct {
	bun.BaseModel `bun:"table:authz_addresses"`

	ID            int64      `bun:",pk,autoincrement"`
	WalletAddress string     `bun:"wallet_address,notnull"`
	CreatedAt     time.Time  `bun:"created_at"`
	DeletedAt     *time.Time `bun:"deleted_at"`
	Permission    string     `bun:"permission"`
}
