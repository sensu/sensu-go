package types

import "context"

// Ring is the interface of a ring. Ring's methods are atomic and
// goroutine-safe.
type Ring interface {
	// Add adds an item to the ring. It returns a non-nil error if the
	// operation failed, or the context is cancelled before the operation
	// completed.
	Add(ctx context.Context, value string) error

	// Remove removes an item from the ring. It returns a non-nil error if the
	// operation failed, or the context is cancelled before the operation
	// completed.
	Remove(ctx context.Context, value string) error

	// Next gets the next item in the Ring. The other items in the Ring will
	// all be returned by subsequent calls to Next before this item is
	// returned again. Next returns the selected value, and an error indicating
	// if the operation failed, or if the context was cancelled.
	Next(context.Context) (string, error)
}

// RingGetter provides a way to get a Ring.
type RingGetter interface {
	// GetRing gets a named Ring.
	GetRing(path ...string) Ring
}
