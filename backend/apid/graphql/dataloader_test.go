package graphql

import (
	"context"

	"github.com/graph-gophers/dataloader"
)

func contextWithLoadersNoCache(ctx context.Context, cfg ServiceConfig, opts ...dataloader.Option) context.Context {
	opts = append(opts, dataloader.WithCache(&dataloader.NoCache{}))
	return contextWithLoaders(ctx, cfg, opts...)
}
