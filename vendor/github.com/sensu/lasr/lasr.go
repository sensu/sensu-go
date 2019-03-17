package lasr

import (
	"fmt"
	"sync"
	"time"

	bolt "github.com/coreos/bbolt"
)

// Q is a persistent message queue. Its methods are goroutine-safe.
// Q retains the data that is sent to it until messages are acked (or nacked
// without retry)
type Q struct {
	db          *bolt.DB
	name        []byte
	seq         Sequencer
	keys        bucketKeys
	messages    *fifo
	closed      chan struct{}
	inFlight    sync.WaitGroup
	waker       *waker
	optsApplied bool
	mu          sync.RWMutex
}

type bucketKeys struct {
	ready     []byte
	returned  []byte
	unacked   []byte
	delayed   []byte
	waiting   []byte
	blockedOn []byte
	blocking  []byte
}

// Close closes q. When q is closed, Send, Receive, and Close will return
// ErrQClosed. Close blocks until all messages in the "unacked" state are Acked
// or Nacked.
func (q *Q) Close() error {
	q.messages.Lock()
	defer q.messages.Unlock()
	select {
	case <-q.closed:
		return ErrQClosed
	default:
		close(q.closed)
	}
	q.inFlight.Wait()
	return q.equilibrate()
}

func (q *Q) isClosed() bool {
	select {
	case <-q.closed:
		return true
	default:
		return false
	}
}

func (q *Q) String() string {
	return fmt.Sprintf("Q{Name: %q}", string(q.name))
}

// NewQ creates a new Q. Only one queue should be created in a given bolt db,
// unless compaction is disabled.
func NewQ(db *bolt.DB, name string, options ...Option) (*Q, error) {
	bName := []byte(name)
	closed := make(chan struct{})
	q := &Q{
		db:   db,
		name: bName,
		keys: bucketKeys{
			ready:     []byte("ready"),
			unacked:   []byte("unacked"),
			delayed:   []byte("delayed"),
			waiting:   []byte("waiting"),
			blockedOn: []byte("blockedOn"),
			blocking:  []byte("blocking"),
		},
		waker:  newWaker(closed),
		closed: closed,
	}
	for _, o := range options {
		if err := o(q); err != nil {
			return nil, fmt.Errorf("lasr: couldn't create Q: %s", err)
		}
	}
	q.optsApplied = true
	if err := q.init(); err != nil {
		return nil, err
	}
	return q, nil
}

func (q *Q) init() error {
	if q.messages == nil {
		q.messages = newFifo(1)
	}
	return q.equilibrate()
}

func (q *Q) equilibrate() error {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.db.Update(func(tx *bolt.Tx) error {
		bucket, err := q.bucket(tx, q.keys.ready)
		if err != nil {
			return err
		}
		readyKeys := bucket.Stats().KeyN
		unacked, err := q.bucket(tx, q.keys.unacked)
		if err != nil {
			return err
		}
		cursor := unacked.Cursor()
		// put unacked messages from previous session back in the queue
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			if err := bucket.Put(k, v); err != nil {
				return err
			}
			readyKeys++
		}
		if readyKeys > 0 && !q.isClosed() {
			q.waker.Wake()
		}
		q.messages.Drain()
		root, err := tx.CreateBucketIfNotExists(q.name)
		if err != nil {
			return err
		}
		if len(q.keys.delayed) > 0 && !q.isClosed() {
			// WakeAt for all delayed messages
			delayed, err := q.bucket(tx, q.keys.delayed)
			if err != nil {
				return err
			}
			delayC := delayed.Cursor()
			for k, _ := delayC.First(); k != nil; k, _ = delayC.Next() {
				var id Uint64ID
				if err := id.UnmarshalBinary(k); err != nil {
					return fmt.Errorf("error reading delayed key %v: %s", k, err)
				}
				q.waker.WakeAt(time.Unix(0, int64(id)))
			}
		}
		// Delete the unacked bucket now that the unacked messages have been
		// returned to the ready bucket.
		return root.DeleteBucket(q.keys.unacked)
	})
}

type bucketer interface {
	CreateBucketIfNotExists([]byte) (*bolt.Bucket, error)
	Bucket([]byte) *bolt.Bucket
}

func (q *Q) bucket(tx *bolt.Tx, key []byte) (*bolt.Bucket, error) {
	bucket, err := tx.CreateBucketIfNotExists(q.name)
	if err != nil {
		return nil, err
	}
	bucket, err = bucket.CreateBucketIfNotExists(key)
	if err != nil {
		return nil, err
	}
	return bucket, nil
}
