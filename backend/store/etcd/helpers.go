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
// for querying multiple elements accross namespaces.
func query(ctx context.Context, store *Store, fn getObjectsPath) (*clientv3.GetResponse, error) {
	var namespace string

	// Support "*" as a wildcard
	if ns := types.ContextNamespace(ctx); ns != types.NamespaceTypeAll {
		namespace = ns
	}

	// Determine if we need to query across multiple namespaces
	if namespace == "" {
		ctx = context.WithValue(ctx, types.NamespaceKey, "")
	}

	resp, err := store.client.Get(ctx, fn(ctx, ""), clientv3.WithPrefix())
	if err != nil {
		return resp, err
	}

	return resp, err
}
