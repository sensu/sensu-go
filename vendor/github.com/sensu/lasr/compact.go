package lasr

import (
	"fmt"
	"os"
	"path/filepath"

	bolt "go.etcd.io/bbolt"
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
	tempPath := filepath.Join(filepath.Dir(q.db.Path()), ".lasr.temp.db")
	newDB, err := bolt.Open(tempPath, 0600, nil)
	if err != nil {
		return fmt.Errorf("error compacting queue: %s", err)
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	if err := compact(newDB, q.db); err != nil {
		return fmt.Errorf("error compacting queue: %s", err)
	}
	dbPath := q.db.Path()
	if err := q.db.Close(); err != nil {
		return fmt.Errorf("error compacting queue: %s", err)
	}
	if err := newDB.Close(); err != nil {
		return fmt.Errorf("error compacting queue: %s", err)
	}
	if err := os.Rename(tempPath, dbPath); err != nil {
		return fmt.Errorf("error compacting queue: %s", err)
	}
	q.db, err = bolt.Open(dbPath, 0644, nil)
	if err != nil {
		return fmt.Errorf("error compacting queue: %s", err)
	}
	return nil
}

// walk walks recursively the bolt database db, calling walkFn for each key it finds.
// copied from https://github.com/boltdb/bolt/blob/master/cmd/bolt/main.go#L1693
func walk(db *bolt.DB, walkFn walkFunc) error {
	return db.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, b *bolt.Bucket) error {
			return walkBucket(b, nil, name, nil, b.Sequence(), walkFn)
		})
	})
}

// copied from https://github.com/boltdb/bolt/blob/master/cmd/bolt/main.go#L1702
func walkBucket(b *bolt.Bucket, keypath [][]byte, k, v []byte, seq uint64, fn walkFunc) error {
	// Execute callback.
	if err := fn(keypath, k, v, seq); err != nil {
		return err
	}

	// If this is not a bucket then stop.
	if v != nil {
		return nil
	}

	// Iterate over each child key/value.
	keypath = append(keypath, k)
	return b.ForEach(func(k, v []byte) error {
		if v == nil {
			bkt := b.Bucket(k)
			return walkBucket(bkt, keypath, k, nil, bkt.Sequence(), fn)
		}
		return walkBucket(b, keypath, k, v, b.Sequence(), fn)
	})
}

type walkFunc func([][]byte, []byte, []byte, uint64) error

func compact(dst, src *bolt.DB) error {
	// 1 MB per transaction
	var txMaxSize int64 = 1048576

	// copied with minor modifications from boltdb repo (https://github.com/boltdb/bolt/blob/master/cmd/bolt/main.go#L1620)
	// commit regularly, or we'll run out of memory for large datasets if using one transaction.
	var size int64
	tx, err := dst.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := walk(src, func(keys [][]byte, k, v []byte, seq uint64) error {
		// On each key/value, check if we have exceeded tx size.
		sz := int64(len(k) + len(v))
		if size+sz > txMaxSize {
			// Commit previous transaction.
			if err := tx.Commit(); err != nil {
				return err
			}

			// Start new transaction.
			tx, err = dst.Begin(true)
			if err != nil {
				return err
			}
			size = 0
		}
		size += sz

		// Create bucket on the root transaction if this is the first level.
		nk := len(keys)
		if nk == 0 {
			bkt, err := tx.CreateBucket(k)
			if err != nil {
				return err
			}
			if err := bkt.SetSequence(seq); err != nil {
				return err
			}
			return nil
		}

		// Create buckets on subsequent levels, if necessary.
		b := tx.Bucket(keys[0])
		if nk > 1 {
			for _, k := range keys[1:] {
				b = b.Bucket(k)
			}
		}

		// If there is no value then this is a bucket call.
		if v == nil {
			bkt, err := b.CreateBucket(k)
			if err != nil {
				return err
			}
			if err := bkt.SetSequence(seq); err != nil {
				return err
			}
			return nil
		}

		// Otherwise treat it as a key/value pair.
		return b.Put(k, v)
	}); err != nil {
		return err
	}

	return tx.Commit()
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
