package leader

import (
	"context"
)

// Do executes f if and only if this node is the leader. It returns the error
// that f returns, or any error it encountered while trying to establish
// leadership.
//
// Typically, functions executed by Do are long-lived and run until they are
// terminated by loss of leadership.
//
// f MUST exit early if its context is cancelled before f is finished. If f
// does not terminate when its context is cancelled, then its actions may be
// concurrent with the next elected leader's.
//
// If this node is not the leader, Do will block until it is elected. This can
// be terminated by calling Resign.
//
// Do can lead to etcd leader elections, which may also fail. These failures
// will be returned as errors.
//
// Do will run f in its own goroutine after coordinating the work, but blocks
// as if the function was being executed synchronously.
func Do(f func(context.Context) error) error {
	if override {
		return f(context.Background())
	}
	if super == nil {
		// Init not called.
		return ErrNotInitialized
	}
	work := newWork(f)
	super.Exec(work)
	return work.Err()
}
