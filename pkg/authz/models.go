package authz

import (
	"time"

	"github.com/uptrace/bun"
)

type IPAddress struct {
	bun.BaseModel `bun:"table:ip_addresses"`

	ID         int64      `bun:",pk,autoincrement"`
	IPAddress  string     `bun:"ip_address,notnull"`
	CreatedAt  time.Time  `bun:"created_at"`
	DeletedAt  *time.Time `bun:"deleted_at"`
	Permission string     `bun:"permission"`
	Comment    string     `bun:"comment"`
}
