package lasr

import "errors"

// If dead-lettering is enabled on q, DeadLetters will return a dead-letter
// queue that is named the same as q, but will emit dead-letters on Receive.
// The dead-letter queue itself does not support dead-lettering; nacked
// messages that are not retried will be deleted.
//
// If dead-lettering is not enabled on q, an error will be returned.
func DeadLetters(q *Q) (*Q, error) {
	if len(q.keys.returned) == 0 {
		return nil, errors.New("lasr: dead-letters not available")
	}
	closed := make(chan struct{})
	d := &Q{
		db:   q.db,
		name: q.name,
		seq:  q.seq,
		keys: bucketKeys{
			ready:   q.keys.returned,
			unacked: []byte("deadletters-unacked"),
		},
		waker:  newWaker(closed),
		closed: closed,
	}
	if err := d.init(); err != nil {
		return nil, err
	}
	return d, nil
}
