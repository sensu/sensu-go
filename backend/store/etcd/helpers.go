package etcd

import (
	"context"

	"github.com/sensu/sensu-go/types"
)

// environment returns the environment name injected in the context
func environment(ctx context.Context) string {
	if value := ctx.Value(types.EnvironmentKey); value != nil {
		return value.(string)
	}
	return ""
}

// organization returns the organization name injected in the context
func organization(ctx context.Context) string {
	if value := ctx.Value(types.OrganizationKey); value != nil {
		return value.(string)
	}
	return ""
}
