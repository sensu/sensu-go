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
	"github.com/sensu/sensu-go/backend/store"
)

var ringPathPrefix = "rings"

var ringKeyBuilder = store.NewKeyBuilder(ringPathPrefix)

// ErrEmptyRing is returned when the ring has no items to retrieve.
var ErrEmptyRing = errors.New("empty ring")

// Interface is the interface of a Ring. Ring's methods are atomic and
// goroutine-safe.
type Interface interface {
	// Add adds an item to the ring. It returns a non-nil error if the
	// operation failed, or the context is cancelled before the operation
	// completed.
	Add(ctx context.Context, value string) error

	// Remove removes an item from the ring. It returns a non-nil error if the
	// operation failed, or the context is cancelled before the operation
	// completed.
	Remove(ctx context.Context, value string) error

	// Next gets the next item in the Ring. The other items in the Ring will
	// all be returned by subsequent calls to Next before this item is
	// returned again. Next returns the selected value, and an error indicating
	// if the operation failed, or if the context was cancelled.
	Next(context.Context) (string, error)

	// Peek gets the next item in the Ring, but does not advance the iteration.
	Peek(context.Context) (string, error)
}

// Getter provides a way to get a Ring.
type Getter interface {
	// GetRing gets a named Ring.
	GetRing(path ...string) Interface
}

// EtcdGetter is an Etcd implementation of Getter.
type EtcdGetter struct {
	*clientv3.Client
}

// GetRing gets a named Ring.
func (e EtcdGetter) GetRing(path ...string) Interface {
	return New(ringKeyBuilder.Build(path...), e.Client)
}

// Ring is a ring of values. Users can cycle through the values in the Ring
// with the Next method. Values can be added or removed from the Ring with Add
// and Remove.
type Ring struct {
	// The name of the ring.
	Name string

	client *clientv3.Client
	kv     clientv3.KV
}

// New returns a new Ring.
func New(name string, client *clientv3.Client) *Ring {
	return &Ring{
		Name:   name,
		client: client,
		kv:     clientv3.NewKV(client),
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

// Add adds a new value to the ring, if it is not already present.
func (r *Ring) Add(ctx context.Context, value string) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		key := newKey(r.Name)
		putCmp := clientv3.Compare(clientv3.Version(key), "=", 0)
		putOp := clientv3.OpPut(key, value)
		cmps, ops, err := r.getRemovalOps(ctx, value)
		if err != nil {
			return err
		}
		cmps = append(cmps, putCmp)
		ops = append(ops, putOp)
		response, err := r.kv.Txn(ctx).If(cmps...).Then(ops...).Commit()
		if err != nil {
			return err
		}
		if response.Succeeded {
			return nil
		}
	}
}

// Remove removes a value from the ring.
func (r *Ring) Remove(ctx context.Context, value string) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		cmps, ops, err := r.getRemovalOps(ctx, value)
		if err != nil {
			return err
		}
		if len(ops) == 0 {
			return nil
		}
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

// Next returns the next item in the Ring and advances the iteration. If the
// Ring is empty, then Next will return an empty value, and ErrEmptyRing.
func (r *Ring) Next(ctx context.Context) (string, error) {
	for {
		response, err := r.client.Get(ctx, r.Name, clientv3.WithFirstKey()...)
		if err != nil {
			return "", err
		}
		if len(response.Kvs) == 0 {
			return "", ErrEmptyRing
		}
		kvs := response.Kvs[0]
		key := string(kvs.Key)
		value := string(kvs.Value)
		delCmp := clientv3.Compare(clientv3.ModRevision(key), "=", kvs.ModRevision)
		delOp := clientv3.OpDelete(key)

		nextKey := newKey(r.Name)
		putCmp := clientv3.Compare(clientv3.ModRevision(nextKey), "=", 0)
		putOp := clientv3.OpPut(nextKey, value)

		resp, err := r.kv.Txn(ctx).If(delCmp, putCmp).Then(delOp, putOp).Commit()
		if err != nil {
			return "", err
		}
		if resp.Succeeded {
			return value, nil
		}
	}
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
