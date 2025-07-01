package server

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

func NewPGXDB(dsn string, waitForDB, statementTimeout time.Duration) (*sql.DB, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	config.ConnConfig.RuntimeParams["statement_timeout"] = fmt.Sprint(statementTimeout.Milliseconds())

	dbpool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, err
	}
	db := stdlib.OpenDBFromPool(dbpool)

	err = WaitUntilDBReady(context.Background(), dbpool, waitForDB)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func newBunPGXDb(dsn string, waitForDB, statementTimeout time.Duration) (*bun.DB, error) {
	pgxDb, err := NewPGXDB(dsn, waitForDB, statementTimeout)
	if err != nil {
		return nil, err
	}

	return bun.NewDB(pgxDb, pgdialect.New()), nil
}

func WaitUntilDBReady(ctx context.Context, db *pgxpool.Pool, waitTime time.Duration) error {
	pingCtx, cancel := context.WithTimeout(ctx, waitTime)
	defer cancel()

	err := db.Ping(pingCtx)
	if err != nil {
		return fmt.Errorf("database is not ready within %s: %w", waitTime, err)
	}
	return nil
}
