package context

import (
	"context"

	"github.com/sensu/sensu-go/types"
)

// Environment returns the environment name injected in the context
func Environment(ctx context.Context) string {
	if value := ctx.Value(types.EnvironmentKey); value != nil {
		return value.(string)
	}
	return ""
}

// Organization returns the organization name injected in the context
func Organization(ctx context.Context) string {
	if value := ctx.Value(types.OrganizationKey); value != nil {
		return value.(string)
	}
	return ""
}
