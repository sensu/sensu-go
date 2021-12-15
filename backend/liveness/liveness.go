package liveness

import (
	"context"
	"fmt"
	"math"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/sensu/sensu-go/backend/etcd"
	"github.com/sensu/sensu-go/backend/store/etcd/kvc"
	"github.com/sirupsen/logrus"
	"go.etcd.io/etcd/api/v3/mvccpb"
	rpctypes "go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/time/rate"
)

// SwitchPrefix contains the base path for switchset, which are tracked under
// path.Join(SwitchPrefix, toggleName, key)
var SwitchPrefix = "/sensu.io/switchsets"

// State represents a custom int type for the key stae
type State int

const (
	// FallbackTTL represents the minimal supported etcd lease TTL,  in case the
	// system encounters a toggle that does not store a TTL
	FallbackTTL = 5

	// Alive state is 0
	Alive State = 0

	// Dead state is 1
	Dead State = 1

	// If a key is marked as buried, it is slated to be deleted
	buried = "buried"
)

func (s State) String() string {
	switch s {
	case Alive:
		return "alive"
	case Dead:
		return "dead"
	default:
		return fmt.Sprintf("invalid<%d>", s)
	}
}

// Interface specifies the interface for liveness
type Interface interface {
	// Alive is an assertion that an entity is alive.
	Alive(ctx context.Context, id string, ttl int64) error

	// Dead is an assertion that an entity is dead. Dead is useful for
	// registering entities that are known to be dead, but not yet tracked.
	Dead(ctx context.Context, id string, ttl int64) error

	// Bury forgets an entity exists
	Bury(ctx context.Context, id string) error

	// BuryAndRevokeLease forgets an entity exists and revokes the lease associated with it
	BuryAndRevokeLease(ctx context.Context, id string) error
}

// Factory is a function that can deliver an Interface
type Factory func(name string, dead, alive EventFunc, logger logrus.FieldLogger) Interface

// EtcdFactory returns a Factory that uses an etcd client. The Interface is
// cached after the first instantiation, and the EventFuncs and logger cannot
// be changed later.
func EtcdFactory(ctx context.Context, client *clientv3.Client) Factory {
	switches := make(map[string]Interface)
	switchMu := new(sync.Mutex)
	return Factory(func(name string, dead, alive EventFunc, logger logrus.FieldLogger) Interface {
		switchMu.Lock()
		defer switchMu.Unlock()
		_, ok := switches[name]
		if !ok {
			ss := NewSwitchSet(client, name, dead, alive, logger)
			ss.monitor(ctx)
			switches[name] = ss
		}
		return switches[name]
	})
}

// SwitchSet is a set of switches that get flipped on life and death events
// for entities. On life and death events, callback functions that are
// registered on NewSwitchSet are started as new goroutines.
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

	// This channel serializes events so that their execution ordering is
	// as expected, without causing undue blocking in the main monitoring
	// loop.
	events chan func() (key string, bury bool)
}

// EventFunc is a function that can be used by a SwitchSet to handle events.
// The previous state of the switch will be passed to the function.
//
// For "dead" EventFuncs, the leader flag can be used to determine if the
// client that flipped the switch is our client. For "alive" EventFuncs,
// this parameter is always false.
//
// The EventFunc should return whether or not to bury the switch. If bury is
// true, then the key associated with the EventFunc will be buried and no
// further events will occur for this key.
type EventFunc func(key string, prev State, leader bool) (bury bool)

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
		events:      make(chan func() (string, bool), 512),
	}
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
func (t *SwitchSet) Alive(ctx context.Context, id string, ttl int64) error {
	return kvc.Backoff(ctx).Retry(func(n int) (done bool, err error) {
		return kvc.RetryRequest(n, t.ping(ctx, id, ttl, true))
	})
}

// BuryAndRevokeLease is similar to Bury() but will revoke the lease associated
// with the switch if one exists.
func (t *SwitchSet) BuryAndRevokeLease(ctx context.Context, id string) error {
	key := path.Join(t.prefix, id)

	// find lease ID that we will revoke later on
	var leaseID clientv3.LeaseID
	var getResp *clientv3.GetResponse
	if err := kvc.Backoff(ctx).Retry(func(n int) (done bool, err error) {
		getResp, err = t.client.Get(ctx, key, clientv3.WithIgnoreValue())
		return kvc.RetryRequest(n, err)
	}); err != nil {
		return fmt.Errorf("switch retrieval failed: %s", err)
	}
	if len(getResp.Kvs) != 0 {
		leaseID = clientv3.LeaseID(getResp.Kvs[0].Lease)
	}

	// bury the key as usual
	if err := t.Bury(ctx, id); err != nil {
		return err
	}

	// revoke the lease from earlier if it exists
	if leaseID != 0 {
		if _, err := t.client.Revoke(ctx, leaseID); err != nil {
			t.logger.Debugf(
				"error revoking lease for buried switch, lease may have already been revoked: %s",
				err)
		}
	}

	return nil
}

// Bury buries a live or dead switch. The switch will no longer
// or callbacks.
func (t *SwitchSet) Bury(ctx context.Context, id string) error {
	key := path.Join(t.prefix, id)

	t.logger.WithFields(logrus.Fields{"key": key}).Debug("burying key")

	// set the value of the switch key to buried
	if err := kvc.Backoff(ctx).Retry(func(n int) (done bool, err error) {
		_, err = t.client.Put(ctx, key, buried)
		return kvc.RetryRequest(n, err)
	}); err != nil {
		t.logger.WithFields(logrus.Fields{"key": key}).Errorf("error burying key: %s", err)
		return fmt.Errorf("error burying switch: %s", err)
	}

	// delete the switch key
	if err := kvc.Backoff(ctx).Retry(func(n int) (done bool, err error) {
		_, err = t.client.Delete(ctx, key)
		return kvc.RetryRequest(n, err)
	}); err != nil {
		t.logger.WithFields(logrus.Fields{"key": key}).Errorf("error deleting key: %s", err)
		return fmt.Errorf("error burying switch: %s", err)
	}

	return nil
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
func (t *SwitchSet) Dead(ctx context.Context, id string, ttl int64) error {
	return kvc.Backoff(ctx).Retry(func(n int) (done bool, err error) {
		return kvc.RetryRequest(n, t.ping(ctx, id, ttl, false))
	})
}

func isBuried(event *clientv3.Event) bool {
	if event.Kv != nil && len(event.Kv.Value) > 0 {
		return string(event.Kv.Value) == buried
	}
	if event.PrevKv != nil {
		return string(event.PrevKv.Value) == buried
	}
	return false
}

func (t *SwitchSet) ping(ctx context.Context, id string, ttl int64, alive bool) error {
	if ttl < FallbackTTL {
		return fmt.Errorf("bad ttl: %d is less than the minimum value of %d", ttl, FallbackTTL)
	}

	putVal := ttl
	if !alive {
		putVal = -putVal
	}

	key := path.Join(t.prefix, id)
	val := fmt.Sprintf("%d", putVal)

	return kvc.Backoff(ctx).Retry(func(n int) (done bool, err error) {
		// WithIgnoreLease will re-use the existing lease
		resp, err := t.client.Put(ctx, key, val, clientv3.WithIgnoreLease(), clientv3.WithPrevKV())
		if err != nil {
			if err.Error() == rpctypes.ErrLeaseNotFound.Error() || err.Error() == rpctypes.ErrKeyNotFound.Error() {
				// The existing lease wasn't found, it must have expired or
				// been revoked. This isn't strictly an error as it can occur
				// in the course of normal operation. As such, we won't track
				// metrics for it.
				//
				// it's ugly, but this is how etcd itself matches this error,
				// so doing it here too.
				leaseID, err := t.newLease(ctx, ttl)
				if err != nil {
					etcd.LeaseOperationsCounter.WithLabelValues("liveness", etcd.LeaseOperationTypePut, etcd.LeaseStatusFor(err)).Inc()
					return kvc.RetryRequest(n, err)
				}
				_, err = t.client.Put(ctx, key, val, clientv3.WithLease(leaseID))
				if err != nil {
					etcd.LeaseOperationsCounter.WithLabelValues("liveness", etcd.LeaseOperationTypePut, etcd.LeaseStatusFor(err)).Inc()
				}
				return kvc.RetryRequest(n, err)
			}
			etcd.LeaseOperationsCounter.WithLabelValues("liveness", etcd.LeaseOperationTypePut, etcd.LeaseStatusFor(err)).Inc()
			return kvc.RetryRequest(n, err)
		}
		leaseID := clientv3.LeaseID(resp.PrevKv.Lease)
		_, err = t.client.KeepAliveOnce(ctx, leaseID)
		return kvc.RetryRequest(n, err)
	})
}

func (t *SwitchSet) getLeaseIDFromKV(ctx context.Context, kv *mvccpb.KeyValue, ttl int64) (clientv3.LeaseID, error) {
	leaseID := clientv3.LeaseID(kv.Lease)
	if leaseID == 0 {
		return t.newLease(ctx, ttl)
	}
	if _, err := t.client.KeepAliveOnce(ctx, leaseID); err != nil {
		etcd.LeaseOperationsCounter.WithLabelValues("liveness", etcd.LeaseOperationTypeKeepalive, etcd.LeaseOperationStatusExpired).Inc()
		return t.newLease(ctx, ttl)
	}
	return leaseID, nil
}

func (t *SwitchSet) newLease(ctx context.Context, ttl int64) (clientv3.LeaseID, error) {
	var leaseID clientv3.LeaseID
	err := kvc.Backoff(ctx).Retry(func(n int) (done bool, err error) {
		lease, err := t.client.Grant(ctx, ttl)
		etcd.LeaseOperationsCounter.WithLabelValues("liveness", etcd.LeaseOperationTypeGrant, etcd.LeaseStatusFor(err)).Inc()
		if err == nil {
			leaseID = lease.ID
		}
		return kvc.RetryRequest(n, err)
	})
	return leaseID, err
}

func (t *SwitchSet) getTTLFromEvent(event *clientv3.Event) (int64, State) {
	var (
		ttl, prev int64
		prevState State
	)
	if event.PrevKv != nil && len(event.PrevKv.Value) > 0 {
		fmt.Sscanf(string(event.PrevKv.Value), "%d", &prev)
	}
	if prev > 0 {
		prevState = Alive
	} else {
		prevState = Dead
	}
	if len(event.Kv.Value) > 0 {
		// A put has resulted in this event, and the TTL is stored here
		fmt.Sscanf(string(event.Kv.Value), "%d", &ttl)
		return ttl, prevState
	}
	if event.PrevKv != nil && len(event.PrevKv.Value) > 0 {
		// The previous revision contains the TTL
		fmt.Sscanf(string(event.PrevKv.Value), "%d", &ttl)
		return ttl, prevState
	}
	t.logger.Errorf("using fallback TTL for %q", string(event.Kv.Key))
	return -FallbackTTL, prevState
}

// monitor starts a goroutine that monitors the SwitchSet prefix for key PUT
// and DELETE events.
func (t *SwitchSet) monitor(ctx context.Context) {
	wc := t.client.Watch(ctx, t.prefix, clientv3.WithPrefix(), clientv3.WithPrevKV())
	go func() {
		for event := range t.events {
			if key, bury := event(); bury {
				id := strings.TrimPrefix(key, t.prefix+"/")
				if err := t.Bury(context.Background(), id); err != nil {
					t.logger.WithError(err).Errorf("error burying %q", key)
				}
			}
		}
	}()
	go func() {
		ctx := clientv3.WithRequireLeader(ctx)
		limiter := rate.NewLimiter(rate.Every(time.Second), 1)
		_ = limiter.Wait(ctx)
	OUTER:
		for {
			select {
			case <-ctx.Done():
				close(t.events)
				return
			case resp, ok := <-wc:
				if err := resp.Err(); err != nil && err != context.Canceled {
					t.logger.WithError(err).Error("error monitoring toggles")
				}
				if resp.Canceled || !ok {
					wc = t.client.Watch(ctx, t.prefix, clientv3.WithPrefix(), clientv3.WithPrevKV())
					goto OUTER
				}
				for _, event := range resp.Events {
					t.handleEvent(ctx, event)
				}
			}
		}
	}()
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
	if isBuried(event) || ctx.Err() != nil {
		// The event was buried - we don't need to handle it, or the context was canceled.
		return
	}
	ttl, prevState := t.getTTLFromEvent(event)
	key := string(event.Kv.Key)

	switch event.Type {
	case mvccpb.DELETE:
		// The entity has expired. Replace it with a new entity
		// to keep the events firing
		t.logger.WithFields(logrus.Fields{"key": key, "ttl": ttl}).Debug("key expired")

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

		t.logger.Debugf("creating a lease for %s with TTL %d", key, leaseTTL)
		leaseID, err := t.getLeaseIDFromKV(ctx, event.Kv, leaseTTL)
		if err != nil {
			t.logger.WithError(err).Errorf("error while granting lease for %s", key)
			return
		}

		// Store a negative value for the TTL to indicate that the
		// entity is not alive.
		put := clientv3.OpPut(key, fmt.Sprintf("%d", ttl), clientv3.WithLease(leaseID))
		resp, err := t.client.Txn(ctx).If(cmp).Then(put).Commit()
		etcd.LeaseOperationsCounter.WithLabelValues("liveness", etcd.LeaseOperationTypePut, etcd.LeaseStatusFor(err)).Inc()
		if err != nil {
			t.logger.WithError(err).Errorf("error commiting keepalive tx for %s", key)
			return
		}
		t.events <- func() (string, bool) {
			return key, t.notifyDead(strings.TrimPrefix(key, t.prefix+"/"), prevState, resp.Succeeded)
		}

	case mvccpb.PUT:
		// Watch PUTs to determine if we need to execute a handler for entity
		// liveness.
		if ttl == 0 {
			t.logger.Errorf("bad PUT for %s: TTL is 0", key)
			return
		}
		if ttl > 0 && ttl != math.MaxInt64 {
			// A positive TTL indicates the entity is alive
			t.logger.Debugf("%s alive: %d", key, ttl)
			t.events <- func() (string, bool) {
				return key, t.notifyAlive(strings.TrimPrefix(key, t.prefix+"/"), prevState, false)
			}
		}
	}
}
