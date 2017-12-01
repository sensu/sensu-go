package etcd

import "github.com/sensu/sensu-go/types"

// rejectByEnvironment returns false if any given record is not part of the
// environment configured in the namespace. This method is useful when a
// wildcard was given for the organization but not for the environment.
func rejectByEnvironment(ns namespace) func(types.MultitenantResource) bool {
	// If the namespace is a wildcard never reject
	if ns.EnvIsWildcard() {
		return func(_ types.MultitenantResource) bool {
			return false
		}
	}

	return func(r types.MultitenantResource) bool {
		return r.GetEnvironment() != ns.env
	}
}
