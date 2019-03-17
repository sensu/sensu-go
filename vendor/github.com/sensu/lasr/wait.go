package lasr

import bolt "github.com/coreos/bbolt"

// Wait causes a message to wait for other messages to Ack, before entering the
// Ready state.
//
// When all of the messages Wait is waiting on have been Acked, then the message
// will enter the Ready state.
//
// When there are no messages to wait on, Wait behaves the same as Send.
func (q *Q) Wait(msg []byte, on ...ID) (ID, error) {
	if len(on) < 1 {
		return q.Send(msg)
	}
	var id ID
	q.mu.RLock()
	defer q.mu.RUnlock()
	return id, q.db.Update(func(tx *bolt.Tx) error {
		var err error
		id, err = q.nextSequence(tx)
		if err != nil {
			return err
		}
		idb, err := id.MarshalBinary()
		if err != nil {
			return err
		}
		blockedOn, err := q.bucket(tx, q.keys.blockedOn)
		if err != nil {
			return err
		}
		blockedMsg, err := blockedOn.CreateBucketIfNotExists(idb)
		if err != nil {
			return err
		}
		blocking, err := q.bucket(tx, q.keys.blocking)
		if err != nil {
			return err
		}
		for _, id := range on {
			idc, err := id.MarshalBinary()
			if err != nil {
				return err
			}
			if err := blockedMsg.Put(idc, nil); err != nil {
				return err
			}
			blockerMsg, err := blocking.CreateBucketIfNotExists(idc)
			if err != nil {
				return err
			}
			if err := blockerMsg.Put(idb, nil); err != nil {
				return err
			}
		}
		waiting, err := q.bucket(tx, q.keys.waiting)
		if err != nil {
			return err
		}
		return waiting.Put(idb, msg)
	})
}
