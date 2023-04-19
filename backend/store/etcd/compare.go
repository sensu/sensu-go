package etcd

import (
	v2 "github.com/sensu/core/v2"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func namespaceExistsForResource(r v2.MultitenantResource) clientv3.Cmp {
	key := getNamespacePath(r.GetNamespace())
	return clientv3.Compare(clientv3.Version(key), ">", 0)
}
