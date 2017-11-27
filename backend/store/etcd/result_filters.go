package etcd

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
