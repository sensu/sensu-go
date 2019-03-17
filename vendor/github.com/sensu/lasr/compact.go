package lasr

import (
	"fmt"
	"io/ioutil"
	"os"

	bolt "github.com/coreos/bbolt"
)

// Compact performs compaction on the queue. All messages will be copied from
// the underlying database into another.
//
// The queue is unavailable while compaction is occurring.
//
// Compaction causes the underlying bolt database to be replaced, so callers
// should be aware that any other queues relying on the database may be
// invalidated.
func (q *Q) Compact() (rerr error) {
	tf, err := ioutil.TempFile("", "")
	if err != nil {
		return fmt.Errorf("error compacting queue: %s", err)
	}
	_ = tf.Close()
	newDB, err := bolt.Open(tf.Name(), 0644, nil)
	if err != nil {
		return fmt.Errorf("error compacting queue: %s", err)
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	err = q.db.Update(func(oldTx *bolt.Tx) error {
		return newDB.Update(func(newTx *bolt.Tx) error {
			oldBucket := oldTx.Bucket(q.name)
			if oldBucket != nil {
				newBucket, err := newTx.CreateBucket(q.name)
				if err != nil {
					return err
				}
				return copyNestedBucket(oldBucket, newBucket)
			}
			return nil
		})
	})
	if err != nil {
		return fmt.Errorf("error compacting queue: %s", err)
	}
	dbPath := q.db.Path()
	if err := q.db.Close(); err != nil {
		return fmt.Errorf("error compacting queue: %s", err)
	}
	tempName := newDB.Path()
	if err := newDB.Close(); err != nil {
		return fmt.Errorf("error compacting queue: %s", err)
	}
	if err := os.Rename(tempName, dbPath); err != nil {
		return fmt.Errorf("error compacting queue: %s", err)
	}
	q.db, err = bolt.Open(dbPath, 0644, nil)
	if err != nil {
		return fmt.Errorf("error compacting queue: %s", err)
	}
	return nil
}

func copyNestedBucket(old, new *bolt.Bucket) error {
	return old.ForEach(func(k, v []byte) error {
		if v != nil {
			return new.Put(k, v)
		}
		nestedOld := old.Bucket(k)
		if nestedOld == nil {
			return new.Put(k, v)
		}
		nestedNew, err := new.CreateBucketIfNotExists(k)
		if err != nil {
			return err
		}
		return copyNestedBucket(nestedOld, nestedNew)
	})
}
