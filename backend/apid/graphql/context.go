package graphql

import (
	"context"

	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/types"
)

// setContextFromComponents takes a context and global id components, adds the
// namespace to the context, and returns the updated context.
func setContextFromComponents(ctx context.Context, c globalid.Components) context.Context {
	return contextWithNamespace(ctx, c.Namespace())
}

// contextWithNamespace takes a context and a name adds it to the context, and
// returns the updated context.
func contextWithNamespace(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, types.NamespaceKey, name)
}
