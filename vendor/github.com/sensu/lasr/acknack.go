package lasr

import (
	"sync/atomic"

	bolt "go.etcd.io/bbolt"
)

func (q *Q) ack(id []byte) error {
	q.mu.RLock()
	defer q.mu.RUnlock()
	var wake bool
	err := q.db.Update(func(tx *bolt.Tx) error {
		var err error
		wake, err = q.stopWaitingOn(tx, id)
		if err != nil {
			return err
		}
		bucket, err := q.bucket(tx, q.keys.unacked)
		if err != nil {
			return err
		}
		return bucket.Delete(id)
	})
	if err == nil {
		q.inFlight.Done()
	}
	if wake && !q.isClosed() {
		q.waker.Wake()
	}
	return err
}

func (q *Q) nack(id []byte, retry bool) error {
	q.mu.RLock()
	defer q.mu.RUnlock()
	var wake bool
	err := q.db.Update(func(tx *bolt.Tx) (rerr error) {
		bucket, err := q.bucket(tx, q.keys.unacked)
		if err != nil {
			return err
		}
		if retry {
			wake = true
			val := bucket.Get(id)
			ready, err := q.bucket(tx, q.keys.ready)
			if err != nil {
				return err
			}
			return ready.Put(id, val)
		}
		wake, err = q.stopWaitingOn(tx, id)
		if err != nil {
			return err
		}
		if len(q.keys.returned) > 0 {
			val := bucket.Get(id)
			returned, err := q.bucket(tx, q.keys.returned)
			if err != nil {
				return err
			}
			return returned.Put(id, val)
		}
		return bucket.Delete(id)
	})
	if err != nil {
		return err
	}
	q.inFlight.Done()
	if wake && !q.isClosed() {
		q.waker.Wake()
	}
	return nil
}

// Ack acknowledges successful receipt and processing of the Message.
func (m *Message) Ack() (err error) {
	if !atomic.CompareAndSwapInt32(&m.once, 0, 1) {
		return ErrAckNack
	}
	return m.q.ack(m.ID)
}

// Nack negatively acknowledges successful receipt and processing of the
// Message. If Nack is called with retry True, then the Message will be
// placed back in the queue in its original position.
func (m *Message) Nack(retry bool) (err error) {
	if !atomic.CompareAndSwapInt32(&m.once, 0, 1) {
		return ErrAckNack
	}
	return m.q.nack(m.ID, retry)
}

// stopWaitingOn causes all messages waiting on id to not wait on id.
// If stopWaitingOn finds any messages that were waiting on id that are not
// waiting on any other messages, it will move them to the Ready state.
func (q *Q) stopWaitingOn(tx *bolt.Tx, id []byte) (bool, error) {
	// blocking -> x blocking y
	// blockedOn -> x blocked on y
	wake := false
	blocking, err := q.bucket(tx, q.keys.blocking)
	if err != nil {
		return wake, err
	}
	blocker := blocking.Bucket(id)
	if blocker == nil {
		return wake, nil
	}
	blockedOn, err := q.bucket(tx, q.keys.blockedOn)
	if err != nil {
		return wake, err
	}
	c := blocker.Cursor()
	for k, _ := c.First(); k != nil; k, _ = c.Next() {
		blockedMsg := blockedOn.Bucket(k)
		if blockedMsg == nil {
			continue
		}
		if err := blockedMsg.Delete(id); err != nil {
			return wake, err
		}
		c = blockedMsg.Cursor()
		if k, _ := c.First(); k != nil {
			// There are still messages blocking the release of blockedMsg
			continue
		}
		// at this point, nothing is causing 'blockedMsg' to wait anymore.
		// a message will be placed into the ready state
		if err := blockedOn.DeleteBucket(k); err != nil {
			return wake, err
		}
		waiting, err := q.bucket(tx, q.keys.waiting)
		if err != nil {
			return wake, err
		}
		v := waiting.Get(k)
		ready, err := q.bucket(tx, q.keys.ready)
		if err != nil {
			return wake, err
		}
		if err := ready.Put(k, v); err != nil {
			return wake, err
		}
		wake = true
		if err := waiting.Delete(k); err != nil {
			return wake, err
		}
	}
	return wake, blocking.DeleteBucket(id)
}
