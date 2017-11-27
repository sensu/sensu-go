package types

import "context"

// MultitenantResource is a object that belongs to an organization
type MultitenantResource interface {
	GetOrganization() string
	GetEnvironment() string
}

// SetContextFromResource takes a context and a multi-tenant resource, adds the environment and
// organization to the context, and returns the udpated context
func SetContextFromResource(ctx context.Context, r MultitenantResource) context.Context {
	ctx = context.WithValue(ctx, EnvironmentKey, r.GetEnvironment())
	ctx = context.WithValue(ctx, OrganizationKey, r.GetOrganization())
	return ctx
}
