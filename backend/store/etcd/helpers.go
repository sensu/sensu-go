package etcd

import (
	"context"
	"encoding/json"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/types"
)

// getObjectsPath functions take a context and an object name and return
// the path of these objects in etcd
type getObjectsPath func(context.Context, string) string

// query is a wrapper around etcd Get method, which provides additional support
// for querying multiple elements accross organizations and environments.
// N.B. Even if we only query across organizations, we still need to filter the
// values returned based on their environment afterwards if the objects type
// doesn't contain the environment at the top level of the object
func query(ctx context.Context, store *Store, fn getObjectsPath) (*clientv3.GetResponse, error) {
	// Support "*" as a wildcard
	var org, env string
	if org = organization(ctx); org == "*" {
		org = ""
	}
	if env = environment(ctx); env == "*" {
		env = ""
	}

	// Determine if we need to query across multiple organizations or environments
	if org == "" {
		ctx = context.WithValue(ctx, types.OrganizationKey, "")
		ctx = context.WithValue(ctx, types.EnvironmentKey, "")
	} else if env == "" {
		ctx = context.WithValue(ctx, types.EnvironmentKey, "")
	}

	resp, err := store.client.Get(ctx, fn(ctx, ""), clientv3.WithPrefix())
	if err != nil {
		return resp, err
	}

	// Return all elements if all environments were requested
	if env == "" {
		return resp, nil
	}

	// Filter elements based on their environment
	var value map[string]interface{}
	for i, kv := range resp.Kvs {
		if err := json.Unmarshal(kv.Value, &value); err != nil {
			// We are dealing with unexpected data, just return the raw data
			return resp, nil
		}

		environment, ok := value["environment"].(string)
		if !ok {
			// We are dealing with an unconvential type of objects (e.g. events)
			// so just return all elements
			return resp, nil
		}

		// Make sure we only keep the elements that are member of the specified env
		if environment != env {
			resp.Kvs = append(resp.Kvs[:i], resp.Kvs[i+1:]...)
		}
	}

	return resp, err
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
