package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"reflect"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/backend/store"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
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

// Create the given key with the serialized object.
func Create(ctx context.Context, client *clientv3.Client, key, namespace string, object interface{}) error {
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
	resp, err := client.Txn(ctx).If(comparisons...).Then(req).Else(
		getNamespace(namespace), getKey(key),
	).Commit()
	if err != nil {
		return err
	}
	if !resp.Succeeded {
		// Check if the namespace was missing
		if namespace != "" && len(resp.Responses[0].GetResponseRange().Kvs) == 0 {
			return &store.ErrNamespaceMissing{Namespace: namespace}
		}

		// Check if the key already exists
		if len(resp.Responses[1].GetResponseRange().Kvs) != 0 {
			return &store.ErrAlreadyExists{Key: key}
		}

		// Unknown error
		return &store.ErrInternal{
			Message: fmt.Sprintf("could not create the key %s", key),
		}
	}

	return nil
}

// CreateOrUpdate writes the given key with the serialized object, regarless of
// its current existence
func CreateOrUpdate(ctx context.Context, client *clientv3.Client, key, namespace string, object interface{}) error {
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
	resp, err := client.Txn(ctx).If(comparisons...).Then(req).Commit()
	if err != nil {
		return err
	}
	if !resp.Succeeded {
		// Check if the namespace was missing
		if namespace != "" && len(resp.Responses[0].GetResponseRange().Kvs) == 0 {
			return &store.ErrNamespaceMissing{Namespace: namespace}
		}

		// Unknown error
		return &store.ErrInternal{
			Message: fmt.Sprintf("could not update the key %s", key),
		}
	}

	return nil
}

// Delete the given key
func Delete(ctx context.Context, client *clientv3.Client, key string) error {
	resp, err := client.Delete(ctx, key)
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

// Get retrieves an object with the given key
func Get(ctx context.Context, client *clientv3.Client, key string, object interface{}) error {
	// Fetch the key from the store
	resp, err := client.Get(ctx, key, clientv3.WithLimit(1))
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

// KeyBuilderFn represents a generic key builder function
type KeyBuilderFn func(context.Context, string) string

// List retrieves all keys from storage under the provided prefix key, while
// supporting all namespaces, and deserialize it into objsPtr.
func List(ctx context.Context, client *clientv3.Client, keyBuilder KeyBuilderFn, objsPtr interface{}, pageSize int64, continueToken string) (string, error) {
	// Make sure the interface is a pointer, and that the element at this address
	// is a slice.
	v := reflect.ValueOf(objsPtr)
	if v.Kind() != reflect.Ptr {
		return "", fmt.Errorf("expected pointer, but got %v type", v.Type())
	}
	if v.Elem().Kind() != reflect.Slice {
		return "", fmt.Errorf("expected slice, but got %s", v.Elem().Kind())
	}
	v = v.Elem()

	opts := []clientv3.OpOption{
		clientv3.WithLimit(pageSize),
	}

	key := keyBuilder(ctx, "")
	rangeEnd := clientv3.GetPrefixRangeEnd(key)
	opts = append(opts, clientv3.WithRange(rangeEnd))

	resp, err := client.Get(ctx, path.Join(key, continueToken), opts...)
	if err != nil {
		return "", err
	}

	for _, kv := range resp.Kvs {
		// Decode and append the value to v, which must be a slice.
		obj := reflect.New(v.Type().Elem()).Interface()
		if err := json.Unmarshal(kv.Value, obj); err != nil {
			return "", &store.ErrDecode{Key: key, Err: err}
		}

		v.Set(reflect.Append(v, reflect.ValueOf(obj).Elem()))
	}

	nextContinueToken := ""
	if pageSize != 0 && resp.Count > pageSize {
		lastObject := v.Index(v.Len() - 1).Interface().(corev2.Resource)
		nextContinueToken = computeContinueToken(ctx, lastObject)
	}

	return nextContinueToken, nil
}

// Update a key given with the serialized object.
func Update(ctx context.Context, client *clientv3.Client, key, namespace string, object interface{}) error {
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
	resp, err := client.Txn(ctx).If(comparisons...).Then(req).Else(
		getNamespace(namespace), getKey(key),
	).Commit()
	if err != nil {
		return err
	}
	if !resp.Succeeded {
		// Check if the namespace was missing
		if namespace != "" && len(resp.Responses[0].GetResponseRange().Kvs) == 0 {
			return &store.ErrNamespaceMissing{Namespace: namespace}
		}

		// Check if the key was missing
		if len(resp.Responses[1].GetResponseRange().Kvs) == 0 {
			return &store.ErrNotFound{Key: key}
		}

		// Unknown error
		return &store.ErrInternal{
			Message: fmt.Sprintf("could not update the key %s", key),
		}
	}

	return nil
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

func keyNotFound(key string) clientv3.Cmp {
	return clientv3.Compare(
		clientv3.CreateRevision(key), "=", 0,
	)
}

func namespaceFound(namespace string) clientv3.Cmp {
	return keyFound(getNamespacePath(namespace))
}

func computeContinueToken(ctx context.Context, r corev2.Resource) (token string) {
	objMeta := r.GetObjectMeta()
	queriedNamespace := store.NewNamespaceFromContext(ctx)

	if queriedNamespace == "" {
		token = path.Join(objMeta.Namespace, objMeta.Name) + "\x00"
	} else {
		token = objMeta.Name + "\x00"
	}

	return
}
