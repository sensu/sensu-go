// Package ring implements a ring in etcd.
package ring

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"path"
	"sync"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/sensu/sensu-go/backend/etcd"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

var (
	// ErrEmptyRing is returned when the ring has no items to retrieve.
	ErrEmptyRing = errors.New("ring: empty ring")

	// ErrNotOwner is returned when a client tries to operate on a ring item
	// that it does not have ownership of.
	ErrNotOwner = errors.New("ring: not owner")

	ringPathPrefix = "rings"
	ringKeyBuilder = store.NewKeyBuilder(ringPathPrefix)

	backendID   string
	backendOnce sync.Once

	leaseIDCache = make(map[string]clientv3.LeaseID)
	pkgMu        sync.Mutex

	initialItemKey []byte
)

func init() {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, uint64(0)); err != nil {
		// Should never happen
		panic(err)
	}
	initialItemKey = buf.Bytes()
}

// EtcdGetter is an Etcd implementation of Getter.
type EtcdGetter struct {
	*clientv3.Client
	BackendID string
}

// GetRing gets a named Ring.
func (e EtcdGetter) GetRing(path ...string) types.Ring {
	return New(ringKeyBuilder.Build(path...), e.Client, e.BackendID)
}

// Ring is a ring of values. Users can cycle through the values in the Ring
// with the Next method. Values can be added or removed from the Ring with Add
// and Remove.
type Ring struct {
	// The name of the ring.
	Name string

	client          *clientv3.Client
	kv              clientv3.KV
	backendID       string
	leaseTimeout    int64
	once            sync.Once
	watchChan       clientv3.WatchChan
	wakeup          chan struct{}
	itemPrefix      string
	assertionPrefix string
	nextItemKey     string
	keySeqKey       string
}

// New returns a new Ring.
func New(name string, client *clientv3.Client, backendID string) *Ring {
	pkgMu.Lock()
	defer pkgMu.Unlock()
	ring := &Ring{
		Name:            name,
		client:          client,
		kv:              clientv3.NewKV(client),
		backendID:       backendID,
		leaseTimeout:    120, // 120 seconds
		wakeup:          make(chan struct{}, 1),
		itemPrefix:      path.Join(name, "items"),
		assertionPrefix: path.Join(name, "owner"),
		nextItemKey:     path.Join(name, "next"),
		keySeqKey:       path.Join(name, "seq"),
	}
	ring.initWatcher()
	return ring
}

func (r *Ring) initWatcher() {
	r.watchChan = r.client.Watch(context.Background(), r.nextItemKey, clientv3.WithPrefix())
}

// supervise tries to acquire a mutex on the ring. If it succeeds, r's client
// will be responsible for doing the ring accounting.
func (r *Ring) supervise() error {
	session, err := concurrency.NewSession(r.client)
	if err != nil {
		return err
	}
	mu := concurrency.NewMutex(session, path.Join(r.Name, "mutex"))
	go func() {
		mu.Lock(context.Background())
		defer mu.Unlock(context.Background())
		for range r.wakeup {
			if err := r.advance(); err != nil {
				logger.WithError(err).Error("supervisor error")
			}
		}
	}()
	return nil
}

// advance takes an item from the head of the ring and puts it at the tail.
// it also does a put of the item value to the nextItemKey. This enables
// other clients to be informed of the ring advancing via the watcher.
// If there are no items in the ring then advance will put an empty value
// to the nextItemKey, so that clients will be able to return ErrEmptyRing.
func (r *Ring) advance() error {
	response, err := r.client.Get(context.Background(), r.itemPrefix, clientv3.WithFirstKey()...)
	if err != nil {
		return fmt.Errorf("error getting next ring item: %s", err)
	}
	if len(response.Kvs) == 0 {
		// Put a nil value so that clients will know the ring is empty
		_, err := r.client.Put(context.TODO(), r.nextItemKey, "")
		return err
	}

	kvs := response.Kvs[0]
	key := string(kvs.Key)
	value := string(kvs.Value)

	// Delete the head of the ring
	delCmp := clientv3.Compare(clientv3.ModRevision(key), "=", kvs.ModRevision)
	delOp := clientv3.OpDelete(key)

	// Place the value at the tail of the ring
	seq, err := etcd.Sequence(r.kv, r.keySeqKey)
	if err != nil {
		return fmt.Errorf("error getting next ring item: %s", err)
	}
	nextKey := path.Join(r.itemPrefix, seq)
	putCmp := clientv3.Compare(clientv3.ModRevision(nextKey), "=", 0)
	putOp := clientv3.OpPut(nextKey, value)

	// Execute the transaction
	resp, err := r.kv.Txn(context.Background()).If(delCmp, putCmp).Then(delOp, putOp).Commit()
	if err != nil {
		return fmt.Errorf("error while executing ring transaction: %s", err)
	}
	if !resp.Succeeded {
		// try again, the transaction was likely invalidated by another client
		// adding or removing items to/from the ring.
		return r.advance()
	}

	// Put the key somewhere the watcher can reliably access it
	_, err = r.client.Put(context.TODO(), r.nextItemKey, value)
	return err
}

// gets a lease ID from cache, or creates a new one if it does not exist yet
func (r *Ring) getLeaseID() (clientv3.LeaseID, error) {
	pkgMu.Lock()
	defer pkgMu.Unlock()
	leaseID, ok := leaseIDCache[r.Name]
	if !ok {
		ctx := context.Background()
		lease, err := r.client.Grant(ctx, r.leaseTimeout)
		if err != nil {
			return 0, err
		}
		ch, err := r.client.KeepAlive(ctx, lease.ID)
		if err != nil {
			return 0, err
		}
		<-ch

		leaseIDCache[r.Name] = lease.ID
		leaseID = lease.ID
	}
	return leaseID, nil
}

// Add adds a new owned value to the ring, which is then owned by the client
// that added it. Only the client that added it will be able to retrieve it with
// Next. If the value already exists, ownership will be transferred to the
// client that most recently added it.
func (r *Ring) Add(ctx context.Context, value string) error {
	for {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("couldn't add item to ring: %s", err)
		}
		seq, err := etcd.Sequence(r.kv, r.keySeqKey)
		if err != nil {
			return fmt.Errorf("couldn't add item to ring: %s", err)
		}
		key := path.Join(r.itemPrefix, seq)
		if err != nil {
			return err
		}
		putCmp := clientv3.Compare(clientv3.Version(key), "=", 0)
		leaseID, err := r.getLeaseID()
		if err != nil {
			return err
		}
		putOp := clientv3.OpPut(key, value, clientv3.WithLease(leaseID))
		// If the item is already in there, remove it and add it again
		cmps, ops, err := r.getRemovalOps(ctx, value)
		if err != nil {
			return err
		}
		ownershipAssertion := clientv3.OpPut(
			path.Join(r.assertionPrefix, value), r.backendID, clientv3.WithLease(leaseID))
		cmps = append(cmps, putCmp)
		ops = append(ops, putOp, ownershipAssertion)
		response, err := r.kv.Txn(ctx).If(cmps...).Then(ops...).Commit()
		if err != nil {
			return err
		}
		if response.Succeeded {
			return nil
		}
	}
}

// Remove removes a value from the ring. It must be owned by the client that
// placed it there, or ErrNotOwner will be returned.
func (r *Ring) Remove(ctx context.Context, value string) error {
	for {
		if err := ctx.Err(); err != nil {
			// The context was cancelled by the caller
			return err
		}
		// Check for the ownership assertion first
		resp, err := r.client.Get(ctx, path.Join(r.assertionPrefix, value))
		if err != nil {
			return err
		}
		if len(resp.Kvs) == 0 {
			// The item has already been deleted by another process
			return nil
		}
		// This client is not the owner of the item, so it should not be
		// allowed to delete it.
		if string(resp.Kvs[0].Value) != r.backendID {
			return ErrNotOwner
		}

		// Get the comparisons and operations necessary to remove the item from
		// the ring.
		cmps, ops, err := r.getRemovalOps(ctx, value)
		if err != nil {
			return err
		}
		if len(ops) == 0 {
			return nil
		}

		// Ensure the owner has not changed
		eqCmp := clientv3.Compare(clientv3.Value(path.Join(r.assertionPrefix, value)), "=", r.backendID)

		// Delete the ownership assertion
		delOp := clientv3.OpDelete(path.Join(r.assertionPrefix, value))
		ops = append(ops, delOp)
		cmps = append(cmps, eqCmp)

		// Transactionally delete the item and its ownership assertion
		response, err := r.kv.Txn(ctx).If(cmps...).Then(ops...).Commit()
		if err != nil {
			return err
		}
		if response.Succeeded {
			return nil
		}
	}
}

// getRemovalOps gets the comparisons and operations necessary to
// transactionally remove an item from the ring, including the item's ownership
// assertion.
func (r *Ring) getRemovalOps(ctx context.Context, value string) ([]clientv3.Cmp, []clientv3.Op, error) {
	// Get all the items in the ring
	response, err := r.client.Get(ctx, r.itemPrefix, clientv3.WithPrefix())
	if err != nil {
		return nil, nil, err
	}
	var cmps []clientv3.Cmp
	var ops []clientv3.Op
	// Compare all the values in the ring with the value supplied
	for _, kv := range response.Kvs {
		key := string(kv.Key)
		if string(kv.Value) == value {
			cmp := clientv3.Compare(clientv3.ModRevision(key), "=", kv.ModRevision)
			cmps = append(cmps, cmp)
			op := clientv3.OpDelete(key)
			ops = append(ops, op)
		}
	}
	return cmps, ops, nil
}

// Next returns the next item in the Ring, if the ring's client owns the item,
// and advances the iteration. If the ring's client does not own the item, then
// ErrNotOwner is returned. If the Ring contains no items whatsoever, then
// ErrNoItems will be returned.
func (r *Ring) Next(ctx context.Context) (string, error) {
	var err error
	r.once.Do(func() {
		err = r.supervise()
	})
	if err != nil {
		return "", fmt.Errorf("error initializing ring: %s", err)
	}
	r.wake()
	watchResponse := <-r.watchChan
	if watchResponse.Canceled {
		r.initWatcher()
		return r.Next(ctx)
	}
	if len(watchResponse.Events) == 0 {
		return r.Next(ctx)
	}
	event := watchResponse.Events[0]
	value := string(event.Kv.Value)
	if len(value) == 0 {
		// The zero-length sentinel value informs us that the ring is empty
		return "", ErrEmptyRing
	}
	isOwner, err := r.owns(value)
	if err != nil {
		return "", fmt.Errorf("error checking key ownership: %s: %s", string(value), err)
	}
	if !isOwner {
		return "", ErrNotOwner
	}
	return string(value), nil
}

// tests whether the key is owned by this backend or another
func (r *Ring) owns(key string) (bool, error) {
	resp, err := r.client.Get(context.Background(), path.Join(r.assertionPrefix, key))
	if err != nil {
		return false, err
	}
	if len(resp.Kvs) == 0 {
		return false, nil
	}
	owner := string(resp.Kvs[0].Value)
	return owner == r.backendID, nil
}

// wakeup the supervisor goroutine, in the case that it was able to obtain the
// etcd mutex.
func (r *Ring) wake() {
	select {
	case r.wakeup <- struct{}{}:
	default:
	}
}
