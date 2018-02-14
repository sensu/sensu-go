// Package ring implements a ring in etcd.
package ring

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"path"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/google/uuid"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

var (
	// ErrEmptyRing is returned when the ring has no items to retrieve.
	ErrEmptyRing = errors.New("ring: empty ring")

	// ErrNotOwner is returned when a client tries to operate on a ring item
	// that it does not have ownership of.
	ErrNotOwner = errors.New("ring: not owner")

	backendID      = uuid.New().String()
	ringPathPrefix = "rings"
	ringKeyBuilder = store.NewKeyBuilder(ringPathPrefix)
)

// EtcdGetter is an Etcd implementation of Getter.
type EtcdGetter struct {
	*clientv3.Client
}

// GetRing gets a named Ring.
func (e EtcdGetter) GetRing(path ...string) types.Ring {
	return New(ringKeyBuilder.Build(path...), e.Client)
}

// Ring is a ring of values. Users can cycle through the values in the Ring
// with the Next method. Values can be added or removed from the Ring with Add
// and Remove.
type Ring struct {
	// The name of the ring.
	Name string

	client       *clientv3.Client
	kv           clientv3.KV
	backendID    string
	leaseTimeout int64
}

// New returns a new Ring.
func New(name string, client *clientv3.Client) *Ring {
	return &Ring{
		Name:         name,
		client:       client,
		kv:           clientv3.NewKV(client),
		backendID:    backendID,
		leaseTimeout: 15, // 15 seconds
	}
}

func newKey(prefix string) string {
	now := time.Now().UnixNano()
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, now); err != nil {
		// Should never happen
		panic(err)
	}
	return path.Join(prefix, buf.String())
}

// Add adds a new owned value to the ring, which is associated with the client
// that added it. Only the client that added it will be able to retrieve it with
// Next. If the value already exists, ownership will be transferred.
func (r *Ring) Add(ctx context.Context, value string) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		lease, err := r.client.Grant(ctx, r.leaseTimeout)
		if err != nil {
			return err
		}
		key := newKey(r.Name)
		putCmp := clientv3.Compare(clientv3.Version(key), "=", 0)
		putOp := clientv3.OpPut(key, value, clientv3.WithLease(lease.ID))
		cmps, ops, err := r.getRemovalOps(ctx, value)
		if err != nil {
			return err
		}
		ownershipAssertion := clientv3.OpPut(
			path.Join(r.Name, value), r.backendID, clientv3.WithLease(lease.ID))
		cmps = append(cmps, putCmp)
		ops = append(ops, putOp, ownershipAssertion)
		response, err := r.kv.Txn(ctx).If(cmps...).Then(ops...).Commit()
		if err != nil {
			return err
		}
		if response.Succeeded {
			ch, err := r.client.KeepAlive(ctx, lease.ID)
			if err != nil {
				return err
			}
			<-ch
			return nil
		}
		if _, err := r.client.Revoke(ctx, lease.ID); err != nil {
			return err
		}
	}
}

// Remove removes a value from the ring. It must be owned by the client that
// placed it there, or ErrNotOwner will be returned.
func (r *Ring) Remove(ctx context.Context, value string) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		resp, err := r.client.Get(ctx, path.Join(r.Name, value))
		if err != nil {
			return err
		}
		if len(resp.Kvs) == 0 {
			return nil
		}
		if string(resp.Kvs[0].Value) != r.backendID {
			return ErrNotOwner
		}
		cmps, ops, err := r.getRemovalOps(ctx, value)
		if err != nil {
			return err
		}
		if len(ops) == 0 {
			return nil
		}
		// Ensure the owner has not changed
		eqCmp := clientv3.Compare(clientv3.Value(path.Join(r.Name, value)), "=", r.backendID)
		// Delete the ownership assertion
		delOp := clientv3.OpDelete(path.Join(r.Name, value))
		ops = append(ops, delOp)
		cmps = append(cmps, eqCmp)
		response, err := r.kv.Txn(ctx).If(cmps...).Then(ops...).Commit()
		if err != nil {
			return err
		}
		if response.Succeeded {
			return nil
		}
	}
}

func (r *Ring) getRemovalOps(ctx context.Context, value string) ([]clientv3.Cmp, []clientv3.Op, error) {
	// Get all the items in the ring
	response, err := r.client.Get(ctx, r.Name, clientv3.WithPrefix())
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

// Next returns the next owned item in the Ring and advances the iteration. If
// the Ring does not contain any items owned by this client, then Next will
// block until the context is cancelled. If the Ring contains no items
// whatsoever, then ErrEmptyRing will be returned.
func (r *Ring) Next(ctx context.Context) (string, error) {
	watchCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	delChan := make(chan struct{})
	wc := r.client.Watch(watchCtx, r.Name, clientv3.WithPrefix())
	if wc == nil {
		return "", ctx.Err()
	}
	go func() {
		// This goroutine watches the ring for delete events
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		defer func() {
			close(delChan)
		}()
		for {
			select {
			case resp := <-wc:
				for _, evt := range resp.Events {
					if evt.Type == mvccpb.DELETE {
						delChan <- struct{}{}
					}
				}
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Use polling as well in case the delete event didn't fire.
				// This seems to happen from time to time, but is very hard
				// to reproduce.
				// TODO: find some way of avoiding this.
				delChan <- struct{}{}
			}
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}
		response, err := r.client.Get(ctx, r.Name, clientv3.WithFirstKey()...)
		if err != nil {
			return "", err
		}
		if len(response.Kvs) == 0 {
			return "", ErrEmptyRing
		}
		ownedKvs, err := r.getOwnedKvs(ctx, response.Kvs)
		if err != nil {
			return "", err
		}
		if len(ownedKvs) == 0 || ownedKvs[0] != response.Kvs[0] {
			<-delChan
			continue
		}
		kvs := response.Kvs[0]
		key := string(kvs.Key)
		value := string(kvs.Value)

		// Ensure ownership has not changed in the meantime
		eqCmp := clientv3.Compare(clientv3.Value(path.Join(r.Name, value)), "=", r.backendID)

		// Delete the head of the ring
		delCmp := clientv3.Compare(clientv3.ModRevision(key), "=", kvs.ModRevision)
		delOp := clientv3.OpDelete(key)

		// Place the value at the tail of the ring
		nextKey := newKey(r.Name)
		putCmp := clientv3.Compare(clientv3.ModRevision(nextKey), "=", 0)
		putOp := clientv3.OpPut(nextKey, value)

		resp, err := r.kv.Txn(ctx).If(delCmp, putCmp, eqCmp).Then(delOp, putOp).Commit()
		if err != nil {
			return "", err
		}
		if resp.Succeeded {
			return value, nil
		}
	}
}

func (r *Ring) getOwnedKvs(ctx context.Context, kvs []*mvccpb.KeyValue) ([]*mvccpb.KeyValue, error) {
	result := make([]*mvccpb.KeyValue, 0, len(kvs))
	for _, kv := range kvs {
		resp, err := r.client.Get(ctx, path.Join(r.Name, string(kv.Value)))
		if err != nil {
			return nil, err
		}
		if len(resp.Kvs) == 0 {
			continue
		}
		if string(resp.Kvs[0].Value) == r.backendID {
			result = append(result, kv)
		}
	}
	return result, nil
}

// Peek returns the next item in the Ring without advancing the iteration. If
// the Ring is empty, then Peek returns an empty string and ErrEmptyRing.
func (r *Ring) Peek(ctx context.Context) (string, error) {
	response, err := r.client.Get(ctx, r.Name, clientv3.WithFirstKey()...)
	if err != nil {
		return "", err
	}
	if len(response.Kvs) == 0 {
		return "", ErrEmptyRing
	}
	return string(response.Kvs[0].Value), nil
}
