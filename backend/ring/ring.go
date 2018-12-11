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
	"github.com/coreos/etcd/mvcc/mvccpb"
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
	ClientID string
}

// GetRing gets a named Ring.
func (e EtcdGetter) GetRing(path ...string) types.Ring {
	return New(ringKeyBuilder.Build(path...), e.Client, e.ClientID)
}

// Ring is a ring of values. Users can cycle through the values in the Ring
// with the Next method. Values can be added or removed from the Ring with Add
// and Remove.
type Ring struct {
	// The name of the ring.
	Name string

	client          *clientv3.Client
	kv              clientv3.KV
	clientID        string
	leaseTimeout    int64
	once            sync.Once
	watchChan       clientv3.WatchChan
	wakeup          chan struct{}
	itemPrefix      string
	assertionPrefix string
	keySeqKey       string
	clientIterKey   string
	ringIterKey     string
}

// New returns a new Ring.
func New(name string, client *clientv3.Client, clientID string) *Ring {
	pkgMu.Lock()
	defer pkgMu.Unlock()
	ring := &Ring{
		Name:            name,
		client:          client,
		kv:              clientv3.NewKV(client),
		clientID:        clientID,
		leaseTimeout:    120, // 120 seconds
		wakeup:          make(chan struct{}, 1),
		itemPrefix:      path.Join(name, "items"),
		assertionPrefix: path.Join(name, "owner"),
		keySeqKey:       path.Join(name, "seq"),
		clientIterKey:   path.Join(name, "iters", clientID),
		ringIterKey:     path.Join(name, "ringiter"),
	}
	return ring
}

// advance takes an item from the head of the ring and puts it at the tail.
func (r *Ring) advance(kv *mvccpb.KeyValue) (uint64, error) {
	key := string(kv.Key)
	value := string(kv.Value)

	// Delete the head of the ring
	delCmp := clientv3.Compare(clientv3.ModRevision(key), "=", kv.ModRevision)
	delOp := clientv3.OpDelete(key)

	// Place the value at the tail of the ring
	seq, err := etcd.Sequence(r.kv, r.keySeqKey)
	if err != nil {
		return 0, fmt.Errorf("error getting next ring item: %s", err)
	}
	nextKey := path.Join(r.itemPrefix, seq)
	putCmp := clientv3.Compare(clientv3.ModRevision(nextKey), "=", 0)
	putOp := clientv3.OpPut(nextKey, value)

	// Execute the transaction
	resp, err := r.kv.Txn(context.Background()).If(delCmp, putCmp).Then(delOp, putOp).Commit()
	if err != nil {
		return 0, fmt.Errorf("error while executing ring transaction: %s", err)
	}
	if !resp.Succeeded {
		// Somehow the transaction failed. Return ErrNotOwner in the hopes that
		// the next call to advance() succeeds. This may cause the ring to miss
		// an iteration, but is possibly the best we can do here.
		return 0, ErrNotOwner
	}

	// Write this to advance the ring iteration #
	ringIter, err := etcd.SequenceUint64(r.kv, r.ringIterKey)
	if err != nil {
		return 0, fmt.Errorf("error while advancing ring: %s", err)
	}

	return ringIter, nil
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
			path.Join(r.assertionPrefix, value), r.clientID, clientv3.WithLease(leaseID))
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
			return fmt.Errorf("couldn't remove item from ring: %s", err)
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
		if string(resp.Kvs[0].Value) != r.clientID {
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
		eqCmp := clientv3.Compare(clientv3.Value(path.Join(r.assertionPrefix, value)), "=", r.clientID)

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

func (r *Ring) clientVersionCurrent(ctx context.Context) (bool, error) {
	ops := []clientv3.Op{
		clientv3.OpGet(r.clientIterKey),
		clientv3.OpGet(r.ringIterKey),
	}
	txn, err := r.client.Txn(ctx).Then(ops...).Commit()
	if err != nil {
		return false, err
	}
	client := txn.Responses[0].GetResponseRange()
	ring := txn.Responses[1].GetResponseRange()
	if len(ring.Kvs) == 0 && len(client.Kvs) == 0 {
		// The ring client has never had Next called, and the ring has never
		// advanced
		return true, nil
	}
	if len(ring.Kvs) != 0 && len(client.Kvs) == 0 {
		// The ring has advanced but the client has never had Next called
		_, err := r.client.Put(ctx, r.clientIterKey, string(ring.Kvs[0].Value))
		return false, err
	}
	current := bytes.Equal(ring.Kvs[0].Value, client.Kvs[0].Value)
	if current {
		return current, nil
	}

	// Make the version current for next time
	_, err = r.client.Put(ctx, r.clientIterKey, string(ring.Kvs[0].Value))
	return current, err
}

// Next returns the next item in the Ring, if the ring's client owns the item,
// and advances the iteration. If the ring's client does not own the item, then
// ErrNotOwner is returned. If the Ring contains no items whatsoever, then
// ErrNoItems will be returned.
func (r *Ring) Next(ctx context.Context) (result string, err error) {
	// If upToDate is false, then the client has missed previous Ring
	upToDate, err := r.clientVersionCurrent(ctx)
	if err != nil {
		return "", fmt.Errorf("error getting next ring item: %s", err)
	}

	if !upToDate {
		// The client calling Next is behind the other clients and the ring
		// has progressed without it. Return ErrNotOwner without taking
		// further action. The client's version will get bumped.
		return "", ErrNotOwner
	}

	// At this point we need to know if we own the next ring item
	resp, err := r.client.Get(ctx, r.itemPrefix, clientv3.WithFirstKey()...)
	if err != nil {
		return "", fmt.Errorf("error getting next ring item: %s", err)
	}
	// There are no ring items
	if len(resp.Kvs) == 0 {
		return "", ErrEmptyRing
	}
	value := resp.Kvs[0].Value
	isOwner, err := r.owns(ctx, string(value))
	if err != nil {
		return "", fmt.Errorf("error checking key ownership: %s: %s", string(value), err)
	}
	if !isOwner {
		// If we don't own the next value, we need to wait until the owner
		// retrieves it and advances the ring before allowing Next to return.
		// Watch for the item to be deleted (moved to the tail).
		watchCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		<-r.client.Watch(watchCtx, string(resp.Kvs[0].Key))
		return "", ErrNotOwner
	}

	// Otherwise, this client does own the item and will be responsible for
	// making the changes to the ring.
	seq, err := r.advance(resp.Kvs[0])
	if err != nil {
		return "", err
	}

	return string(value), etcd.SetSequence(r.kv, r.clientIterKey, seq)
}

// tests whether the key is owned by this backend or another
func (r *Ring) owns(ctx context.Context, key string) (bool, error) {
	resp, err := r.client.Get(ctx, path.Join(r.assertionPrefix, key))
	if err != nil {
		return false, err
	}
	if len(resp.Kvs) == 0 {
		return false, nil
	}
	owner := string(resp.Kvs[0].Value)
	return owner == r.clientID, nil
}
