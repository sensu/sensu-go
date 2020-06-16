package etcdstore

import (
	"context"
	"path"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/proto"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/backend/store/v2/wrap"
	"github.com/sensu/sensu-go/util/retry"
)

const (
	// EtcdRoot is the root of all sensu storage.
	EtcdRoot = "/sensu.io"

	// EtcdInitialDelay is 100 ms.
	EtcdInitialDelay = time.Millisecond * 100
)

var (
	namespaceStoreName                   = new(corev2.Namespace).StorePrefix()
	_                  storev2.Interface = new(Store)
)

// ComputeContinueToken calculates a continue token based on the given resource.
// The resource can be a core/v2 or a core/v3 resource.
func ComputeContinueToken(namespace string, r interface{}) string {
	switch resource := r.(type) {
	case *corev2.Event:
		// TODO(ccressent): This can surely be simplified
		if namespace == "" {
			// Workaround for sensu-go#2465: keepalive events do not always have
			// their namespace filled in, which would break the construction of
			// continue token below. To accommodate for that, when
			// constructing the continue token, whevener an event has a
			// namespace of "" we construct the continue token using its
			// entity's namespace instead.
			eventNamespace := resource.Namespace
			if eventNamespace == "" {
				eventNamespace = resource.Entity.Namespace
			}
			return "/" + eventNamespace + "/" + resource.Entity.Name + "/" + resource.Check.Name + "\x00"
		}
		return resource.Entity.Name + "/" + resource.Check.Name + "\x00"

	case *corev2.Namespace:
		return resource.Name + "\x00"

	case *corev2.User:
		return resource.Username + "\x00"

	case corev2.Resource:
		objMeta := resource.GetObjectMeta()

		if namespace == "" {
			return path.Join(objMeta.Namespace, objMeta.Name) + "\x00"
		}
		return objMeta.Name + "\x00"
	case corev3.Resource:
		objMeta := resource.GetMetadata()

		if namespace == "" {
			return path.Join(objMeta.Namespace, objMeta.Name) + "\x00"
		}
		return objMeta.Name + "\x00"
	default:
		return "invalid-continue-token"
	}
}

func getKey(key string) clientv3.Op {
	return clientv3.OpGet(key)
}

func getNamespace(namespace string) clientv3.Op {
	return getKey(getNamespacePath(namespace))
}

func keyFound(key string) clientv3.Cmp {
	return clientv3.Compare(
		clientv3.CreateRevision(key), ">", 0,
	)
}

func getNamespacePath(name string) string {
	return path.Join(EtcdRoot, namespaceStoreName, name)
}

func keyNotFound(key string) clientv3.Cmp {
	return clientv3.Compare(
		clientv3.CreateRevision(key), "=", 0,
	)
}

func namespaceFound(namespace string) clientv3.Cmp {
	return keyFound(getNamespacePath(namespace))
}

// StoreKey converts a ResourceRequest into a key that uniquely identifies a
// singular resource, or collection of resources, in a namespace.
func StoreKey(req storev2.ResourceRequest) string {
	return store.NewKeyBuilder(req.StoreName).WithNamespace(req.Namespace).Build(req.Name)
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

// Backoff delivers a pre-configured backoff object, suitable for use in making
// etcd requests.
func Backoff(ctx context.Context) *retry.ExponentialBackoff {
	return &retry.ExponentialBackoff{
		Ctx:                  ctx,
		InitialDelayInterval: EtcdInitialDelay,
	}
}

// RetryRequest will return whether or not to try a request again based on the
// error given to it, and the number of times the request has been tried.
//
// If RetryRequest gets "etcdserver: too many requests", then it will return
// (false, nil). Otherwise, it will return (true, err).
func RetryRequest(n int, err error) (bool, error) {
	if err == nil {
		return true, nil
	}
	if err == context.Canceled {
		return true, err
	}
	if err == context.DeadlineExceeded {
		return true, err
	}
	// using string comparison here because it's too difficult to tell
	// what kind of error the client is actually delivering
	if strings.Contains(err.Error(), "etcdserver: too many requests") {
		logger.WithError(err).WithField("retry", n).Error("retrying")
		return false, nil
	}
	return true, &store.ErrInternal{Message: err.Error()}
}

func (s *Store) CreateOrUpdate(req storev2.ResourceRequest, w *storev2.Wrapper) error {
	key := StoreKey(req)
	if err := req.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}
	ctx := req.Context
	msg, err := proto.Marshal(w)
	if err != nil {
		return &store.ErrEncode{Key: key, Err: err}
	}
	comparisons := []clientv3.Cmp{}
	// If we had a namespace provided, make sure it exists
	if req.Namespace != "" {
		comparisons = append(comparisons, namespaceFound(req.Namespace))
	}
	op := clientv3.OpPut(StoreKey(req), string(msg))
	var resp *clientv3.TxnResponse
	err = Backoff(ctx).Retry(func(n int) (done bool, err error) {
		resp, err = s.client.Txn(ctx).If(comparisons...).Then(op).Commit()
		return RetryRequest(n, err)
	})
	if err != nil {
		return err
	}
	if !resp.Succeeded {
		if req.Namespace != "" {
			return &store.ErrNamespaceMissing{Namespace: req.Namespace}
		}

		// should never happen, developer error!
		return &store.ErrInternal{
			Message: "developer error: no namespace specified, but transaction failed",
		}
	}
	return nil
}

func (s *Store) UpdateIfExists(req storev2.ResourceRequest, w *storev2.Wrapper) error {
	key := StoreKey(req)
	if err := req.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}
	ctx := req.Context
	msg, err := proto.Marshal(w)
	if err != nil {
		return &store.ErrEncode{Key: key, Err: err}
	}
	comparisons := []clientv3.Cmp{keyFound(key)}
	op := clientv3.OpPut(StoreKey(req), string(msg))
	var resp *clientv3.TxnResponse
	err = Backoff(ctx).Retry(func(n int) (done bool, err error) {
		resp, err = s.client.Txn(ctx).If(comparisons...).Then(op).Commit()
		return RetryRequest(n, err)
	})
	if err != nil {
		return err
	}
	if !resp.Succeeded {
		return &store.ErrNotFound{Key: key}
	}
	return nil
}

func (s *Store) CreateIfNotExists(req storev2.ResourceRequest, w *storev2.Wrapper) error {
	key := StoreKey(req)
	if err := req.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}
	ctx := req.Context
	msg, err := proto.Marshal(w)
	if err != nil {
		return &store.ErrEncode{Key: key, Err: err}
	}
	comparisons := []clientv3.Cmp{}
	if req.Namespace != "" {
		comparisons = append(comparisons, namespaceFound(req.Namespace))
	}
	comparisons = append(comparisons, keyNotFound(key))
	op := clientv3.OpPut(StoreKey(req), string(msg))
	elseOps := []clientv3.Op{}
	if req.Namespace != "" {
		op := clientv3.OpGet(getNamespacePath(req.Namespace), clientv3.WithCountOnly(), clientv3.WithLimit(1))
		elseOps = append(elseOps, op)
	}
	var resp *clientv3.TxnResponse
	err = Backoff(ctx).Retry(func(n int) (done bool, err error) {
		resp, err = s.client.Txn(ctx).If(comparisons...).Then(op).Else(elseOps...).Commit()
		return RetryRequest(n, err)
	})
	if err != nil {
		return err
	}
	if !resp.Succeeded {
		if len(resp.Responses) > 0 && resp.Responses[0].GetResponseRange().Count == 0 {
			return &store.ErrNamespaceMissing{Namespace: req.Namespace}
		}
		return &store.ErrAlreadyExists{Key: key}
	}
	return nil
}

func (s *Store) Get(req storev2.ResourceRequest) (*storev2.Wrapper, error) {
	key := StoreKey(req)
	if err := req.Validate(); err != nil {
		return nil, &store.ErrNotValid{Err: err}
	}
	ctx := req.Context
	var resp *clientv3.GetResponse
	err := Backoff(ctx).Retry(func(n int) (done bool, err error) {
		resp, err = s.client.Get(ctx, key, clientv3.WithLimit(1))
		return RetryRequest(n, err)
	})
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, &store.ErrNotFound{Key: key}
	}
	var wrapper storev2.Wrapper
	if err := proto.UnmarshalMerge(resp.Kvs[0].Value, &wrapper); err != nil {
		return nil, &store.ErrDecode{Key: key, Err: err}
	}
	return &wrapper, nil
}

func (s *Store) Delete(req storev2.ResourceRequest) error {
	key := StoreKey(req)
	if err := req.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}
	ctx := req.Context
	cmp := keyFound(key)
	op := clientv3.OpDelete(key)
	var resp *clientv3.TxnResponse
	err := Backoff(ctx).Retry(func(n int) (done bool, err error) {
		resp, err = s.client.Txn(ctx).If(cmp).Then(op).Commit()
		return RetryRequest(n, err)
	})
	if err != nil {
		return err
	}
	if !resp.Succeeded {
		return &store.ErrNotFound{Key: key}
	}
	return nil
}

func (s *Store) List(req storev2.ResourceRequest, pred *store.SelectionPredicate) (wrap.List, error) {
	// For any list request, the name must be zeroed out so that the key can
	// be correctly generated.
	req.Name = ""
	key := StoreKey(req)
	if err := req.Validate(); err != nil {
		return nil, &store.ErrNotValid{Err: err}
	}
	ctx := req.Context
	if pred == nil {
		pred = &store.SelectionPredicate{}
	}
	opts := []clientv3.OpOption{
		clientv3.WithLimit(pred.Limit),
		clientv3.WithSerializable(),
	}
	rangeEnd := clientv3.GetPrefixRangeEnd(key)
	opts = append(opts, clientv3.WithRange(rangeEnd))

	if pred.Continue != "" {
		key = path.Join(key, pred.Continue)
	} else {
		if !strings.HasSuffix(key, "/") {
			key += "/"
		}
	}

	var resp *clientv3.GetResponse
	err := Backoff(ctx).Retry(func(n int) (done bool, err error) {
		resp, err = s.client.Get(ctx, key, opts...)
		return RetryRequest(n, err)
	})
	if err != nil {
		return nil, err
	}

	result := make([]*storev2.Wrapper, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		var wrapper storev2.Wrapper
		if err := proto.Unmarshal(kv.Value, &wrapper); err != nil {
			return nil, &store.ErrDecode{Key: string(kv.Key), Err: err}
		}
		result = append(result, &wrapper)
	}
	if pred.Limit != 0 && resp.Count > pred.Limit {
		lastObj, err := result[len(result)-1].Unwrap()
		if err != nil {
			return nil, &store.ErrDecode{Key: string(resp.Kvs[len(resp.Kvs)-1].Key), Err: err}
		}
		pred.Continue = ComputeContinueToken(req.Namespace, lastObj)
	} else {
		pred.Continue = ""
	}
	return result, nil
}
