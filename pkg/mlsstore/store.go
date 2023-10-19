package mlsstore

import (
	"context"

	"github.com/uptrace/bun"
	"go.uber.org/zap"
)

type Store struct {
	ctx    context.Context
	cancel context.CancelFunc
	log    *zap.Logger
	db     *bun.DB
}

type MlsStore interface {
}

func New(config Config) (*Store, error) {
	s := &Store{
		log: config.Log.Named("mlsstore"),
		db:  config.DB,
	}
	s.ctx, s.cancel = context.WithCancel(context.Background())

	return s, nil
}
