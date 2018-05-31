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
	if org = types.ContextOrganization(ctx); org == "*" {
		org = ""
	}
	if env = types.ContextEnvironment(ctx); env == "*" {
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

	// Return all elements if all environments or assets were requested
	if env == "" {
		return resp, nil
	}

	// Filter elements based on their environment
	var value struct {
		Environment *string `json:"environment"`
	}

	for i, kv := range resp.Kvs {
		if err := json.Unmarshal(kv.Value, &value); err != nil {
			// We are dealing with unexpected data, just return the raw data
			return resp, nil
		}

		// Check for the existence of the environment key
		if value.Environment == nil {
			// We are dealing with an unconvential type of objects (e.g. events)
			// so just return all elements
			return resp, nil
		}

		// Make sure we only keep the elements that are member of the specified env
		if *value.Environment != env {
			resp.Kvs = append(resp.Kvs[:i], resp.Kvs[i+1:]...)
		}
	}

	return resp, err
}
