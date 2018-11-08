package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"reflect"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

const (
	// EtcdRoot is the root of all sensu storage.
	EtcdRoot = "/sensu.io"
)

// Store is an implementation of the sensu-go/backend/store.Store iface.
type Store struct {
	client         *clientv3.Client
	keepalivesPath string
}

// NewStore creates a new Store.
func NewStore(client *clientv3.Client, name string) *Store {
	store := &Store{
		client:         client,
		keepalivesPath: path.Join(EtcdRoot, keepalivesPathPrefix, name),
	}

	return store
}

// Create a key given with the serialized object.
func (s *Store) create(ctx context.Context, key, namespace string, object interface{}) error {
	bytes, err := json.Marshal(object)
	if err != nil {
		return &store.ErrEncode{Key: key, Err: err}
	}

	comparisons := []clientv3.Cmp{}
	// If we had a namespace provided, make sure it exists
	if namespace != "" {
		comparisons = append(comparisons, namespaceFound(namespace))
	}
	// Make sure the key does not exists
	comparisons = append(comparisons, keyNotFound(key))

	req := clientv3.OpPut(key, string(bytes))
	resp, err := s.client.Txn(ctx).If(comparisons...).Then(req).Commit()
	if err != nil {
		return err
	}
	if !resp.Succeeded {
		return &store.ErrAlreadyExists{Key: key}
	}

	return nil
}

// CreateOrUpdate writes the given key with the serialized object, regarless of
// its current existence
func (s *Store) createOrUpdate(ctx context.Context, key, namespace string, object interface{}) error {
	bytes, err := json.Marshal(object)
	if err != nil {
		return &store.ErrEncode{Key: key, Err: err}
	}

	comparisons := []clientv3.Cmp{}
	// If we had a namespace provided, make sure it exists
	if namespace != "" {
		comparisons = append(comparisons, namespaceFound(namespace))
	}

	req := clientv3.OpPut(key, string(bytes))
	resp, err := s.client.Txn(ctx).If(comparisons...).Then(req).Commit()
	if err != nil {
		return err
	}
	if !resp.Succeeded {
		return fmt.Errorf("could not create the key %s", key)
	}

	return nil
}

// delete the given key
func (s *Store) delete(ctx context.Context, key string) error {
	resp, err := s.client.Delete(ctx, key)
	if err != nil {
		return err
	}
	if resp.Deleted == 0 {
		return &store.ErrNotFound{Key: key}
	} else if resp.Deleted > 1 {
		return &store.ErrInternal{
			Message: fmt.Sprintf("expected to delete exactly 1 key, deleted %d", resp.Deleted),
		}
	}

	return nil
}

// get retrieves an object with the given key
func (s *Store) get(ctx context.Context, key string, object interface{}) error {
	// Fetch the key from the store
	resp, err := s.client.Get(ctx, key, clientv3.WithLimit(1))
	if err != nil {
		return err
	}

	// Ensure we only received a single item
	if len(resp.Kvs) == 0 {
		return &store.ErrNotFound{Key: key}
	}

	// Deserialize the object to the given object
	if err := json.Unmarshal(resp.Kvs[0].Value, object); err != nil {
		return &store.ErrDecode{Key: key, Err: err}
	}

	return nil
}

// keyBuilderFn represents a generic key builder function
type keyBuilderFn func(context.Context, string) string

// list retrieves all keys from storage under the provided prefix key, while
// supporting all namespaces, and deserialize it into objsPtr.
func (s *Store) list(ctx context.Context, keyBuilder keyBuilderFn, objsPtr interface{}) error {
	// Make sure the interface is a pointer, and that the element at this address
	// is a slice.
	v := reflect.ValueOf(objsPtr)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("expected pointer, but got %v type", v.Type())
	}
	if v.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("expected slice, but got %s", v.Elem().Kind())
	}
	v = v.Elem()

	// Support "*" as a wildcard for namespaces
	if namespace := types.ContextNamespace(ctx); namespace == types.NamespaceTypeAll {
		// Remove the namespace from the context if we had a wildcard
		ctx = context.WithValue(ctx, types.NamespaceKey, "")
	}

	opts := []clientv3.OpOption{
		clientv3.WithPrefix(),
	}

	key := keyBuilder(ctx, "")
	resp, err := s.client.Get(ctx, key, opts...)
	if err != nil {
		return err
	}

	if len(resp.Kvs) == 0 {
		return &store.ErrNotFound{Key: key}
	}

	for _, kv := range resp.Kvs {
		// Decode and append the value to v, which must be a slice.
		obj := reflect.New(v.Type().Elem()).Interface()
		if err := json.Unmarshal(kv.Value, obj); err != nil {
			return &store.ErrDecode{Key: key, Err: err}
		}

		v.Set(reflect.Append(v, reflect.ValueOf(obj).Elem()))
	}

	return nil
}

// Update a key given with the serialized object.
func (s *Store) update(ctx context.Context, key, namespace string, object interface{}) error {
	bytes, err := json.Marshal(object)
	if err != nil {
		return &store.ErrEncode{Key: key, Err: err}
	}

	comparisons := []clientv3.Cmp{}
	// If we had a namespace provided, make sure it exists
	if namespace != "" {
		comparisons = append(comparisons, namespaceFound(namespace))
	}
	// Make sure the key already exists
	comparisons = append(comparisons, keyFound(key))

	req := clientv3.OpPut(key, string(bytes))
	resp, err := s.client.Txn(ctx).If(comparisons...).Then(req).Commit()
	if err != nil {
		return err
	}
	if !resp.Succeeded {
		return fmt.Errorf("could not update the key %s", key)
	}

	return nil
}

func keyFound(key string) clientv3.Cmp {
	return clientv3.Compare(
		clientv3.CreateRevision(key), ">", 0,
	)
}

func keyNotFound(key string) clientv3.Cmp {
	return clientv3.Compare(
		clientv3.CreateRevision(key), "=", 0,
	)
}

func namespaceFound(namespace string) clientv3.Cmp {
	return keyFound(getNamespacePath(namespace))
}
