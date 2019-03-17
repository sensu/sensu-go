package lasr

import (
	"bytes"
	"fmt"
	"time"

	bolt "github.com/coreos/bbolt"
)

var (
	// MaxDelayTime is the maximum time that can be passed to Q.Delay().
	MaxDelayTime = time.Unix(0, 1<<63-1)
)

// Delay is like Send, but the message will not enter the Ready state until
// after "when" has occurred.
//
// If "when" has already occurred, then it will be set to time.Now().
func (q *Q) Delay(message []byte, when time.Time) (ID, error) {
	if when.After(MaxDelayTime) {
		return nil, fmt.Errorf("time out of range: %s", when.Format(time.RFC3339))
	}
	if when.Before(time.Now()) {
		when = time.Now()
	}
	id := Uint64ID(when.UnixNano())
	key, err := id.MarshalBinary()
	if err != nil {
		return nil, err
	}
	err = q.db.Update(func(tx *bolt.Tx) error {
		bucket, err := q.bucket(tx, q.keys.delayed)
		if err != nil {
			return err
		}
		// Reserve a spot for the message. If its exact time in unix
		// nanoseconds has already been reserved, pick the next spot,
		// ad-infinitum.
		for {
			k, _ := bucket.Cursor().Seek(key)
			if !bytes.Equal(k, key) {
				break
			}
			id++
			key, err = id.MarshalBinary()
			if err != nil {
				return err
			}
		}
		return bucket.Put(key, message)
	})
	if err == nil {
		q.waker.WakeAt(time.Unix(0, int64(id)))
	}
	return id, err
}
