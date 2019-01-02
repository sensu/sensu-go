package liveness

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/sirupsen/logrus"
)

// All switchset entities are tracked under path.Join(SwitchPrefix, toggleName, key)
var SwitchPrefix = "/sensu.io/switchsets"

type State int

const (
	// The system has discovered a toggle that did not store a TTL. Use the minimum
	// supported etcd lease TTL as a fallback.
	FallbackTTL = 5

	// The Alive state is 0
	Alive State = 0

	// The Dead state is 1
	Dead State = 1
)

// Interface specifies the interface for liveness
type Interface interface {
	// Alive is an assertion that an entity is alive.
	Alive(ctx context.Context, key string, ttl int64) error

	// Dead is an assertion that an entity is dead. Dead is useful for
	// registering entities that are known to be dead, but not yet tracked.
	Dead(ctx context.Context, key string, ttl int64) error
}

// Factory is a function that can deliver an Interface
type Factory func(name string, dead, alive EventFunc, logger logrus.FieldLogger) Interface

// EtcdFactory returns a Factory that uses an etcd client
func EtcdFactory(client *clientv3.Client) Factory {
	return Factory(func(name string, dead, alive EventFunc, logger logrus.FieldLogger) Interface {
		return NewSwitchSet(client, name, dead, alive, logger)
	})
}

// SwitchSet is a set of switches that get flipped on life and death events
// for entities. On life and death events, callback functions that are
// reigstered on NewSwitchSet are started as new goroutines.
//
// The SwitchSet uses the Alive method to both register members of the set,
// and to assert their liveness once registered. After its first call to
// Alive, if a member does not assert its liveness, then it will be presumed
// to be dead, and the callback for dead members will be called.
//
// When an entity in a SwitchSet dies, it gains a new life in the underworld.
// In the underworld, a dead callback is issued for every TTL interval.
// Entities can go from being dead to alive by calling Alive. When that
// happens, an entity that lives in the underworld will be reborn.
type SwitchSet struct {
	client      *clientv3.Client
	prefix      string
	notifyDead  EventFunc
	notifyAlive EventFunc
	logger      logrus.FieldLogger
}

// EventFunc is a function that can be used by a SwitchSet to handle events.
// The previous state of the switch will be passed to the function, as well as
// the ModRevision of the etcd key. The revision can be used to synchronize
// clients, if need be.
type EventFunc func(key string, prev State, revision int64)

// NewSwitchSet creates a new SwitchSet. It will use an etcd prefix of
// path.Join(SwitchPrefix, name). The dead and live callbacks will be called
// on all life and death events.
func NewSwitchSet(client *clientv3.Client, name string, dead, alive EventFunc, logger logrus.FieldLogger) *SwitchSet {
	return &SwitchSet{
		client:      client,
		prefix:      path.Join(SwitchPrefix, name),
		notifyDead:  dead,
		notifyAlive: alive,
		logger:      logger,
	}
}

func (t *SwitchSet) ping(ctx context.Context, key string, ttl int64, alive bool) error {
	if ttl < FallbackTTL {
		return fmt.Errorf("bad ttl: %d is less than the minimum value of %d", ttl, FallbackTTL)
	}
	if !alive {
		ttl = -ttl
	}
	key = path.Join(t.prefix, key)
	val := fmt.Sprintf("%x", ttl)
	lease, err := t.client.Grant(ctx, ttl)
	if err != nil {
		return err
	}
	_, err = t.client.Put(ctx, key, val, clientv3.WithLease(lease.ID), clientv3.WithPrevKV())
	return err
}

// Alive is an assertion that an entity is alive.
//
// If the SwitchSet doesn't know about the entity yet, then it will be
// registered, and the TTL countdown will start. Unless the entity continually
// asserts its liveness with calls to Alive, it will be presumed dead.
//
// The ttl parameter is the time-to-live in seconds for the entity. The minimum
// TTL value is 5. If a smaller value is passed, then an error will be returned
// and no registration will occur.
func (t *SwitchSet) Alive(ctx context.Context, key string, ttl int64) error {
	return t.ping(ctx, key, ttl, true)
}

// Dead is an assertion that an entity is dead. Dead is useful for registering
// entities that are known to be dead, but not yet tracked by the SwitchSet.
//
// If the SwitchSet doesn't know about the entity yet, then it will be
// registered, and the TTL countdown will start. Until the entity
// asserts its liveness, it will be presumed dead, and dead callbacks will
// be issued.
//
// The ttl parameter is the time-to-live in seconds for the entity. The minimum
// TTL value is 5. If a smaller value is passed, then an error will be returned
// and no registration will occur.
func (t *SwitchSet) Dead(ctx context.Context, key string, ttl int64) error {
	return t.ping(ctx, key, ttl, false)
}

func (t *SwitchSet) getTTLFromEvent(event *clientv3.Event) (int64, State) {
	var (
		ttl, prev int64
		prevState State
	)
	if event.PrevKv != nil && len(event.PrevKv.Value) > 0 {
		fmt.Sscanf(string(event.PrevKv.Value), "%x", &prev)
	}
	if prev > 0 {
		prevState = Alive
	} else {
		prevState = Dead
	}
	if len(event.Kv.Value) > 0 {
		// A put has resulted in this event, and the TTL is stored here
		fmt.Sscanf(string(event.Kv.Value), "%x", &ttl)
		return ttl, prevState
	}
	if event.PrevKv != nil && len(event.PrevKv.Value) > 0 {
		// The previous revision contains the TTL
		fmt.Sscanf(string(event.PrevKv.Value), "%x", &ttl)
		return ttl, prevState
	}
	t.logger.Errorf("using fallback TTL for %q", string(event.Kv.Key))
	return -FallbackTTL, prevState
}

// Monitor starts a goroutine that monitors the SwitchSet prefix for key PUT
// and DELETE events.
func (t *SwitchSet) Monitor(ctx context.Context) {
	wc := t.client.Watch(ctx, t.prefix, clientv3.WithPrefix(), clientv3.WithPrevKV())
	for {
		select {
		case <-ctx.Done():
			return
		case resp := <-wc:
			if err := resp.Err(); err != nil {
				t.logger.WithError(err).Error("error monitoring toggles")
				continue
			}
			for _, event := range resp.Events {
				t.handleEvent(ctx, event)
			}
		}
	}
}

// handleEvent handles a watch event, either DELETE or PUT.
//
// In the case of DELETE, an entity has expired, or an undead entity has been
// replaced. Before any further action is taken, the handler is invoked as a
// goroutine. After, the undead entity is replaced by another undead entity
// with the same undead lifespan. (PUT) In concrete terms, the same key is PUT
// with the same lease TTL as the previous key.
//
// In the case of PUT, the value associated with the event key is checked to
// determine if it is a positive or negative value. If the value is a positive
// value, the entity is either now alive, or still alive. If the value is a
// negative value, then it is ignored, as it is only an undead entity being
// replaced by another undead entity.
func (t *SwitchSet) handleEvent(ctx context.Context, event *clientv3.Event) {
	ttl, prevState := t.getTTLFromEvent(event)
	key := string(event.Kv.Key)
	switch event.Type {
	case mvccpb.DELETE:
		// The entity has expired. Replace it with a new entity
		// to keep the events firing
		go t.notifyDead(strings.TrimPrefix(key, t.prefix+"/"), prevState, event.Kv.ModRevision)
		t.logger.Infof("key expired: %s, ttl: %d", key, ttl)

		// If the key doesn't exist, the version will be 0. This is done to
		// prevent other clients from performing the same operation
		// concurrently.
		cmp := clientv3.Compare(clientv3.Version(key), "=", 0)
		var leaseTTL int64
		if ttl < 0 {
			// In this case, the entity was undead. A negative value is stored
			// to indicate this to the PUT case.
			leaseTTL = -ttl
		} else {
			// In this case, the entity was alive. Put a negative TTL for the
			// value in order to indicate that the entity is now dead.
			leaseTTL = ttl
			ttl = -ttl
		}

		t.logger.Infof("creating a lease for %s with TTL %d", key, leaseTTL)
		lease, err := t.client.Grant(ctx, leaseTTL)
		if err != nil {
			t.logger.WithError(err).Errorf("error while granting lease for %s", key)
			return
		}

		// Store a negative value for the TTL to indicate that the
		// entity is not alive.
		put := clientv3.OpPut(key, fmt.Sprintf("%x", ttl), clientv3.WithLease(lease.ID))
		_, err = t.client.Txn(ctx).If(cmp).Then(put).Commit()
		if err != nil {
			t.logger.WithError(err).Errorf("error commiting keepalive tx for %s", key)
		}

	case mvccpb.PUT:
		// Watch PUTs to determine if we need to execute a handler for entity
		// liveness.
		if ttl == 0 {
			t.logger.Errorf("bad PUT for %s: TTL is 0", key)
			return
		}
		if ttl > 0 {
			// A positive TTL indicates the entity is alive
			t.logger.Infof("%s alive: %d", key, ttl)
			go t.notifyAlive(strings.TrimPrefix(key, t.prefix+"/"), prevState, event.Kv.ModRevision)
		}
	}
}
