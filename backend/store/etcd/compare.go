package etcd

func environmentExistsForResource(r types.MutlitenantResource) {
	key := getEnvironmentsPath(r.GetOrganization(), r.GetEnvironment())
	return clientv3.Compare(clientv3.Version(key), ">", 0)
}
