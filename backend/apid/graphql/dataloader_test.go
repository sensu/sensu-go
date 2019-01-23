package graphql

import (
	"context"

	"github.com/graph-gophers/dataloader"
	"github.com/sensu/sensu-go/cli/client"
)

func contextWithLoadersNoCache(ctx context.Context, client client.APIClient, opts ...dataloader.Option) context.Context {
	opts = append(opts, dataloader.WithCache(&dataloader.NoCache{}))
	return contextWithLoaders(ctx, client, opts...)
}
