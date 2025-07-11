package store

import (
	"context"
	"database/sql"

	"github.com/xmtp/xmtp-node-go/pkg/mls/store/queries"
)

func RunInTx(
	ctx context.Context,
	db *sql.DB,
	opts *sql.TxOptions,
	fn func(ctx context.Context, txQueries *queries.Queries) error,
) error {
	tx, err := db.BeginTx(ctx, opts)
	if err != nil {
		return err
	}

	var done bool

	defer func() {
		if !done {
			_ = tx.Rollback()
		}
	}()

	q := queries.New(tx)

	if err := fn(ctx, q); err != nil {
		return err
	}

	done = true
	return tx.Commit()
}
