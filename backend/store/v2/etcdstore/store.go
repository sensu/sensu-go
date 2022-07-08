package etcdstore

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/gogo/protobuf/proto"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/etcd/kvc"
	"github.com/sensu/sensu-go/backend/store/patch"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/backend/store/v2/wrap"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	_ storev2.Interface = new(Store)
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

func (s *Store) CreateOrUpdate(ctx context.Context, req storev2.ResourceRequest, wrapper storev2.Wrapper) error {
	key := StoreKey(req)
	if err := req.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	w, ok := wrapper.(*wrap.Wrapper)
	if !ok {
		return &store.ErrNotValid{Err: fmt.Errorf("etcdstore only works with wrap.Wrapper, not %T", wrapper)}
	}

	msg, err := proto.Marshal(w)
	if err != nil {
		return &store.ErrEncode{Key: key, Err: err}
	}

	comparator := kvc.Comparisons(
		kvc.NamespaceExists(req.Namespace),
	)
	op := clientv3.OpPut(key, string(msg))

	return kvc.Txn(ctx, s.client, comparator, op)
}

func (s *Store) Patch(ctx context.Context, req storev2.ResourceRequest, wrapper storev2.Wrapper, patcher patch.Patcher, conditions *store.ETagCondition) error {
	if err := req.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	w, ok := wrapper.(*wrap.Wrapper)
	if !ok {
		return &store.ErrNotValid{Err: fmt.Errorf("etcdstore only works with wrap.Wrapper, not %T", wrapper)}
	}

	key := StoreKey(req)

	// Get the stored resource along with the etcd response so we can use the
	// revision later to ensure the resource wasn't modified in the mean time
	resp, err := s.GetWithResponse(ctx, req)
	if err != nil {
		return err
	}
	value := resp.Kvs[0].Value
	if err := proto.UnmarshalMerge(value, w); err != nil {
		return &store.ErrDecode{Key: key, Err: err}
	}

	// Unwrap the stored resource
	resource, err := w.Unwrap()
	if err != nil {
		return &store.ErrDecode{Key: key, Err: err}
	}

	// Now determine the etag for the stored resource
	etag, err := store.ETag(resource)
	if err != nil {
		return err
	}

	if conditions != nil {
		if !store.CheckIfMatch(conditions.IfMatch, etag) {
			return &store.ErrPreconditionFailed{Key: key}
		}
		if !store.CheckIfNoneMatch(conditions.IfNoneMatch, etag) {
			return &store.ErrPreconditionFailed{Key: key}
		}
	}

	// Encode the stored resource to the JSON format
	original, err := json.Marshal(resource)
	if err != nil {
		return err
	}

	// Apply the patch to our original document (stored resource)
	patchedResource, err := patcher.Patch(original)
	if err != nil {
		return err
	}

	// Decode the resulting JSON document back into our resource
	if err := json.Unmarshal(patchedResource, &resource); err != nil {
		return err
	}

	// Validate the resource
	if err := resource.Validate(); err != nil {
		return err
	}

	// Special case for entities; we need to make sure we keep the per-entity
	// subscription
	if e, ok := resource.(*corev3.EntityConfig); ok {
		e.Subscriptions = corev2.AddEntitySubscription(e.Metadata.Name, e.Subscriptions)
	}

	// Re-wrap the resource
	wrappedPatch, err := wrap.Resource(resource)
	if err != nil {
		return &store.ErrEncode{Key: key, Err: err}
	}
	*w = *wrappedPatch

	comparisons := []kvc.Predicate{
		kvc.KeyIsFound(key),
		kvc.KeyHasValue(key, value),
	}

	return s.Update(ctx, req, w, comparisons...)
}

func (s *Store) UpdateIfExists(ctx context.Context, req storev2.ResourceRequest, wrapper storev2.Wrapper) error {
	w, ok := wrapper.(*wrap.Wrapper)
	if !ok {
		return &store.ErrNotValid{Err: fmt.Errorf("etcdstore only works with wrap.Wrapper, not %T", wrapper)}
	}
	key := StoreKey(req)
	comparisons := []kvc.Predicate{
		kvc.NamespaceExists(req.Namespace),
		kvc.KeyIsFound(key),
	}

	return s.Update(ctx, req, w, comparisons...)
}

func (s *Store) Update(ctx context.Context, req storev2.ResourceRequest, wrapper storev2.Wrapper, comparisons ...kvc.Predicate) error {
	w, ok := wrapper.(*wrap.Wrapper)
	if !ok {
		return &store.ErrNotValid{Err: fmt.Errorf("etcdstore only works with wrap.Wrapper, not %T", wrapper)}
	}
	key := StoreKey(req)
	if err := req.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	msg, err := proto.Marshal(w)
	if err != nil {
		return &store.ErrEncode{Key: key, Err: err}
	}

	comparator := kvc.Comparisons(comparisons...)
	op := clientv3.OpPut(key, string(msg))

	return kvc.Txn(ctx, s.client, comparator, op)
}

func (s *Store) CreateIfNotExists(ctx context.Context, req storev2.ResourceRequest, wrapper storev2.Wrapper) error {
	w, ok := wrapper.(*wrap.Wrapper)
	if !ok {
		return &store.ErrNotValid{Err: fmt.Errorf("etcdstore only works with wrap.Wrapper, not %T", wrapper)}
	}
	key := StoreKey(req)
	if err := req.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	msg, err := proto.Marshal(w)
	if err != nil {
		return &store.ErrEncode{Key: key, Err: err}
	}

	comparator := kvc.Comparisons(
		kvc.NamespaceExists(req.Namespace),
		kvc.KeyIsNotFound(key),
	)
	op := clientv3.OpPut(key, string(msg))

	return kvc.Txn(ctx, s.client, comparator, op)
}

func (s *Store) Get(ctx context.Context, req storev2.ResourceRequest) (storev2.Wrapper, error) {
	key := StoreKey(req)
	resp, err := s.GetWithResponse(ctx, req)
	if err != nil {
		return nil, err
	}

	var wrapper wrap.Wrapper
	if err := proto.UnmarshalMerge(resp.Kvs[0].Value, &wrapper); err != nil {
		return nil, &store.ErrDecode{Key: key, Err: err}
	}
	return &wrapper, nil
}

func (s *Store) GetWithResponse(ctx context.Context, req storev2.ResourceRequest) (*clientv3.GetResponse, error) {
	key := StoreKey(req)
	if err := req.Validate(); err != nil {
		return nil, &store.ErrNotValid{Err: err}
	}
	var resp *clientv3.GetResponse
	err := kvc.Backoff(ctx).Retry(func(n int) (done bool, err error) {
		resp, err = s.client.Get(ctx, key, clientv3.WithLimit(1), clientv3.WithSerializable())
		return kvc.RetryRequest(n, err)
	})
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, &store.ErrNotFound{Key: key}
	}

	return resp, nil
}

func (s *Store) Delete(ctx context.Context, req storev2.ResourceRequest) error {
	key := StoreKey(req)
	if err := req.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	comparator := kvc.Comparisons(
		kvc.KeyIsFound(key),
	)
	op := clientv3.OpDelete(key)

	return kvc.Txn(ctx, s.client, comparator, op)
}

func (s *Store) List(ctx context.Context, req storev2.ResourceRequest, pred *store.SelectionPredicate) (storev2.WrapList, error) {
	// For any list request, the name must be zeroed out so that the key can
	// be correctly generated.
	req.Name = ""
	key := StoreKey(req)
	if err := req.Validate(); err != nil {
		return nil, &store.ErrNotValid{Err: err}
	}
	if pred == nil {
		pred = &store.SelectionPredicate{}
	}
	opts := []clientv3.OpOption{
		clientv3.WithLimit(pred.Limit),
		clientv3.WithSerializable(),
		clientv3.WithSort(clientv3.SortByKey, getSortOrder(req.SortOrder)),
	}
	rangeEnd := ""
	if pred.Continue != "" && req.SortOrder == storev2.SortDescend {
		rangeEnd = path.Join(key, strings.TrimRight(pred.Continue, "\x00"))
	} else {
		rangeEnd = clientv3.GetPrefixRangeEnd(key)
	}
	opts = append(opts, clientv3.WithRange(rangeEnd))

	if pred.Continue != "" && req.SortOrder != storev2.SortDescend {
		key = path.Join(key, pred.Continue)
	} else {
		if !strings.HasSuffix(key, "/") {
			key += "/"
		}
	}

	var resp *clientv3.GetResponse
	err := kvc.Backoff(ctx).Retry(func(n int) (done bool, err error) {
		resp, err = s.client.Get(ctx, key, opts...)
		return kvc.RetryRequest(n, err)
	})
	if err != nil {
		return nil, err
	}

	result := make(wrap.List, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		var wrapper wrap.Wrapper
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

func (s *Store) Exists(ctx context.Context, req storev2.ResourceRequest) (bool, error) {
	key := StoreKey(req)
	if err := req.Validate(); err != nil {
		return false, &store.ErrNotValid{Err: err}
	}
	var resp *clientv3.GetResponse
	err := kvc.Backoff(ctx).Retry(func(n int) (done bool, err error) {
		resp, err = s.client.Get(ctx, key, clientv3.WithLimit(1), clientv3.WithSerializable(), clientv3.WithCountOnly())
		return kvc.RetryRequest(n, err)
	})
	if err != nil {
		return false, err
	}
	return resp.Count > 0, nil
}

func getSortOrder(order storev2.SortOrder) clientv3.SortOrder {
	switch order {
	case storev2.SortAscend:
		return clientv3.SortAscend
	case storev2.SortDescend:
		return clientv3.SortDescend
	}
	return clientv3.SortNone
}
