package prune

import "context"

type Pruner interface {
	Count(ctx context.Context) (int64, error)
	PruneCycle(ctx context.Context) (int, error)
	Vacuum(ctx context.Context) error
	Name() string
}
