package lasr

import (
	"bytes"
	"context"
	"time"

	bolt "github.com/coreos/bbolt"
)

// Send sends a message to Q. When send completes with nil error, the message
// sent to Q will be in the Ready state.
func (q *Q) Send(message []byte) (ID, error) {
	if q.isClosed() {
		return nil, ErrQClosed
	}
	var id ID
	q.mu.RLock()
	err := q.db.Update(func(tx *bolt.Tx) (err error) {
		id, err = q.nextSequence(tx)
		if err != nil {
			return err
		}
		return q.send(id, message, tx)
	})
	q.mu.RUnlock()
	if err == nil {
		q.waker.Wake()
	}
	return id, err
}

func (q *Q) send(id ID, body []byte, tx *bolt.Tx) error {
	key, err := id.MarshalBinary()
	if err != nil {
		return err
	}

	bucket, err := q.bucket(tx, q.keys.ready)
	if err != nil {
		return err
	}

	return bucket.Put(key, body)
}

// Receive receives a message from the queue. If no messages are available by
// the time the context is done, then the function will return a nil Message
// and the result of ctx.Err().
func (q *Q) Receive(ctx context.Context) (*Message, error) {
	if q.isClosed() {
		return nil, ErrQClosed
	}
	q.messages.Lock()
	defer q.messages.Unlock()
START:
	if q.messages.Len() > 0 {
		msg := q.messages.Pop()
		err := msg.err
		if err != nil {
			msg = nil
		} else {
			q.inFlight.Add(1)
		}
		return msg, err
	}
	select {
	case <-q.waker.C:
		q.processReceives()
		goto START
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-q.closed:
		return nil, ErrQClosed
	}
}

func (q *Q) processReceives() {
	q.messages.SetError(q.db.Update(func(tx *bolt.Tx) error {
		// Prioritize delayed messages first. Not all instances of Q will
		// have delayed messages.
		if len(q.keys.delayed) > 0 {
			if err := q.getMessages(tx, q.keys.delayed); err != nil {
				return err
			}
			if q.messages.Len() == q.messages.Cap() {
				return nil
			}
		}
		return q.getMessages(tx, q.keys.ready)
	}))
}

func (q *Q) getMessages(tx *bolt.Tx, key []byte) error {
	bucket, err := q.bucket(tx, key)
	if err != nil {
		return err
	}
	cur := bucket.Cursor()
	i := q.messages.Len()
	var currentTime []byte
	if bytes.Equal(key, q.keys.delayed) {
		// special case for processing delays
		t := Uint64ID(time.Now().UnixNano())
		currentTime, err = t.MarshalBinary()
		if err != nil {
			return err
		}
	}
	for k, v := cur.First(); k != nil && i < q.messages.Cap(); k, v = cur.Next() {
		if currentTime != nil && bytes.Compare(k, currentTime) > 0 {
			return nil
		}
		id := cloneBytes(k)
		body := cloneBytes(v)
		unacked, err := q.bucket(tx, q.keys.unacked)
		if err != nil {
			return err
		}
		if err := unacked.Put(k, v); err != nil {
			return err
		}
		if err := bucket.Delete(k); err != nil {
			return err
		}
		q.messages.Push(&Message{
			Body: body,
			ID:   id,
			q:    q,
		})
		i++
	}
	if i >= q.messages.Cap() {
		// More work could be available
		q.waker.Wake()
	}
	return nil
}

func cloneBytes(b []byte) []byte {
	r := make([]byte, len(b))
	copy(r, b)
	return r
}
