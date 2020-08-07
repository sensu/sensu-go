package etcd

import (
	"go.etcd.io/etcd/clientv3"
	"github.com/sensu/sensu-go/types"
)

func namespaceExistsForResource(r types.MultitenantResource) clientv3.Cmp {
	key := getNamespacePath(r.GetNamespace())
	return clientv3.Compare(clientv3.Version(key), ">", 0)
}
