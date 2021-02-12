package etcd

import (
	"github.com/sensu/sensu-go/types"
	"go.etcd.io/etcd/clientv3"
)

func namespaceExistsForResource(r types.MultitenantResource) clientv3.Cmp {
	key := getNamespacePath(r.GetNamespace())
	return clientv3.Compare(clientv3.Version(key), ">", 0)
}
