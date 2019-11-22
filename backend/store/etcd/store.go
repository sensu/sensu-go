package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"reflect"
	"strings"

	"github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/proto"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"

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
func Create(ctx context.Context, client *clientv3.Client, key, namespace string, object proto.Message) error {
	bytes, err := proto.Marshal(object)
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
	var bytes []byte
	var err error

	switch object.(type) {
	case types.Wrapper:
		// Supporting protobuf serialization for wrapped resources is not
		// straightforward since the types.Wrapper struct holds an interface. We
		// will just use JSON encoding for now since the all store functions support
		// both for decoding.
		bytes, err = json.Marshal(object)
		if err != nil {
			return &store.ErrEncode{Key: key, Err: err}
		}
	default:
		msg, ok := object.(proto.Message)
		if !ok {
			return &store.ErrEncode{Key: key, Err: fmt.Errorf("%T is not proto.Message", object)}
		}
		bytes, err = proto.Marshal(msg)
		if err != nil {
			return &store.ErrEncode{Key: key, Err: err}
		}
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
		if namespace != "" && (len(resp.Responses) == 0 || len(resp.Responses[0].GetResponseRange().Kvs) == 0) {
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
	if err := unmarshal(resp.Kvs[0].Value, object); err != nil {
		return &store.ErrDecode{Key: key, Err: err}
	}

	return nil
}

// KeyBuilderFn represents a generic key builder function
type KeyBuilderFn func(context.Context, string) string

// List retrieves all keys from storage under the provided prefix key, while
// supporting all namespaces, and deserialize it into objsPtr.
func List(ctx context.Context, client *clientv3.Client, keyBuilder KeyBuilderFn, objsPtr interface{}, pred *store.SelectionPredicate) error {
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

	opts := []clientv3.OpOption{
		clientv3.WithLimit(pred.Limit),
	}

	keyPrefix := keyBuilder(ctx, "")
	rangeEnd := clientv3.GetPrefixRangeEnd(keyPrefix)
	opts = append(opts, clientv3.WithRange(rangeEnd))

	key := keyPrefix
	if pred.Continue != "" {
		key = path.Join(keyPrefix, pred.Continue)
	} else {
		if !strings.HasSuffix(key, "/") {
			key += "/"
		}
	}

	resp, err := client.Get(ctx, key, opts...)
	if err != nil {
		return err
	}

	for _, kv := range resp.Kvs {
		var obj interface{}
		if len(kv.Value) > 0 && kv.Value[0] == '{' {
			obj = reflect.New(v.Type().Elem().Elem()).Interface()
			if err := json.Unmarshal(kv.Value, obj); err != nil {
				return &store.ErrDecode{Key: key, Err: err}
			}
		} else {
			msg := reflect.New(v.Type().Elem().Elem()).Interface().(proto.Message)
			if err := proto.Unmarshal(kv.Value, msg); err != nil {
				return &store.ErrDecode{Key: key, Err: err}
			}
			obj = msg
		}

		// Initialize the annotations and labels if they are nil
		objValue := reflect.ValueOf(obj)
		if objValue.Kind() == reflect.Ptr {
			meta := objValue.Elem().FieldByName("ObjectMeta")
			if meta.CanSet() {
				if meta.FieldByName("Labels").Len() == 0 && meta.FieldByName("Labels").CanSet() {
					meta.FieldByName("Labels").Set(reflect.MakeMap(reflect.TypeOf(make(map[string]string))))
				}
				if meta.FieldByName("Annotations").Len() == 0 && meta.FieldByName("Annotations").CanSet() {
					meta.FieldByName("Annotations").Set(reflect.MakeMap(reflect.TypeOf(make(map[string]string))))
				}
			}
		}

		v.Set(reflect.Append(v, reflect.ValueOf(obj)))
	}

	if pred.Limit != 0 && resp.Count > pred.Limit {
		lastObject := v.Index(v.Len() - 1).Interface().(corev2.Resource)
		pred.Continue = ComputeContinueToken(ctx, lastObject)
	} else {
		pred.Continue = ""
	}

	return nil
}

// Update a key given with the serialized object.
func Update(ctx context.Context, client *clientv3.Client, key, namespace string, object proto.Message) error {
	bytes, err := proto.Marshal(object)
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

// Count retrieves the count of all keys from storage under the
// provided prefix key, while supporting all namespaces.
func Count(ctx context.Context, client *clientv3.Client, key string) (int64, error) {
	opts := []clientv3.OpOption{
		clientv3.WithCountOnly(),
		clientv3.WithRange(clientv3.GetPrefixRangeEnd(key)),
	}

	resp, err := client.Get(ctx, key, opts...)
	if err != nil {
		return 0, err
	}

	return resp.Count, nil
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

// ComputeContinueToken calculates a continue token based on the given resource
func ComputeContinueToken(ctx context.Context, r corev2.Resource) string {
	queriedNamespace := store.NewNamespaceFromContext(ctx)

	switch resource := r.(type) {
	case *corev2.Event:
		// TODO(ccressent): This can surely be simplified
		if queriedNamespace == "" {
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

	default:
		objMeta := r.GetObjectMeta()

		if queriedNamespace == "" {
			return path.Join(objMeta.Namespace, objMeta.Name) + "\x00"
		}
		return objMeta.Name + "\x00"
	}
}

func unmarshal(data []byte, v interface{}) error {
	if len(data) > 0 && data[0] == '{' {
		if err := json.Unmarshal(data, v); err != nil {
			return err
		}
	} else {
		msg, ok := v.(proto.Message)
		if !ok {
			return fmt.Errorf("%T is not proto.Message", v)
		}
		if err := proto.Unmarshal(data, msg); err != nil {
			return err
		}
	}

	return nil
}
