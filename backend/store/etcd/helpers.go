package etcd

import (
	"context"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/types"
)

// getObjectsPath functions take a context and an object name and return
// the path of these objects in etcd
type getObjectsPath func(context.Context, string) string

// query is a wrapper around etcd Get method, which provides additional support
// for querying multiple elements accross organizations and environments.
// N.B. Even if we only query across organizations, we still need to filter the
// values returned based on their environment afterwards since objects from all
// environments and organizations will be returned
func query(ctx context.Context, store *etcdStore, fn getObjectsPath) (*clientv3.GetResponse, error) {
	org := organization(ctx)
	env := environment(ctx)

	// Determine if we need to query across multiple organizations or environments
	if org == "*" {
		ctx = context.WithValue(ctx, types.OrganizationKey, "")
		ctx = context.WithValue(ctx, types.EnvironmentKey, "")
	} else if env == "*" {
		ctx = context.WithValue(ctx, types.EnvironmentKey, "")
	}

	return store.kvc.Get(ctx, fn(ctx, ""), clientv3.WithPrefix())
}

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
