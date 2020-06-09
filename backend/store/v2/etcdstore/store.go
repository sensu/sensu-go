package etcdstore

import (
	"context"
	"errors"

	"github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/proto"
	storev2 "github.com/sensu/sensu-go/back/store/v2"
	"github.com/sensu/sensu-go/backend/store/wrap"
)

const (
	// EtcdRoot is the root of all sensu storage.
	EtcdRoot = "/sensu.io"
)

func StoreKey(req storev2.ResourceRequest) string {
	return store.NewKeyBuilder(req.StoreSuffix).WithNamespace(req.Namespace).Build(name)
}

// Store is an implementation of the sensu-go/backend/store.Store iface.
type Store struct {
	client *clientv3.Client
}

// NewStore creates a new Store.
func NewStore(client *clientv3.Client) *Store {
	store := &Store{
		client: client,
	}

	return store
}

func (s *Store) CreateOrUpdate(ctx context.Context, req storev2.ResourceRequest, w *wrap.Wrapper) error {
	msg, err := proto.Marshal(w)
	if err != nil {
		return &store.ErrEncode{Key: w.Value.URIPath(), Err: err}
	}
	comparisons := []clientv3.Cmp{}
	// If we had a namespace provided, make sure it exists
	if namespace != "" {
		comparisons = append(comparisons, namespaceFound(req.Namespace))
	}
	op := clientv3.OpPut(StoreKey(req), string(msg))
	var resp *clientv3.TxnResponse
	err = store.Backoff(ctx).Retry(func(n int) (done bool, err error) {
		resp, err = client.Txn(ctx).If(comparisons...).Then(req).Commit()
	})
	if err != nil {
		return err
	}
	if !resp.Succeeded {
		if namespace != "" && (len(resp.Response) == 0 || len(resp.Responses[0].GetResponseRange().Kvs) == 0) {
			return &store.ErrNamespaceMissing{Namespace: req.Namespace}
		}

		return &store.ErrNotValid{
			Err: errors.New("failed to write %s.%s", w.TypeMeta.APIVersion, w.TypeMeta.Type),
		}
	}
	return nil
}
