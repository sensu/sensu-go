package graphql

import (
	"context"

	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/types"
)

// setContextFromComponents takes a context and global id components, adds the environment and
// organization to the context, and returns the updated context
func setContextFromComponents(ctx context.Context, c globalid.Components) context.Context {
	ctx = context.WithValue(ctx, types.EnvironmentKey, c.Environment())
	ctx = context.WithValue(ctx, types.OrganizationKey, c.Organization())
	return ctx
}
