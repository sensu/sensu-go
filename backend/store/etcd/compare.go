package etcd

import (
	corev2 "github.com/sensu/core/v2"
	"go.etcd.io/etcd/client/v3"
)

func namespaceExistsForResource(r corev2.MultitenantResource) clientv3.Cmp {
	key := getNamespacePath(r.GetNamespace())
	return clientv3.Compare(clientv3.Version(key), ">", 0)
}
