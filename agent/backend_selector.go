package agent

import (
	"math/rand"
	"time"
)

// A BackendSelector is repsonsible for selecting an appropriate backend from
// a provided list of backends.
type BackendSelector interface {
	// Select returns an appropriate backend given the selection strategy for
	// the selector.
	Select() string
}

// A RandomBackendSelector does a single random shuffle of a list of backends
// and perpetually returns them in the shuffled order.
//
// RandomBackendSelector is not guaranteed to maintain shuffle order if used by
// multiple goroutines concurrently.
type RandomBackendSelector struct {
	// Backends is the list of backend URLs to shuffle through.
	Backends []string

	shuffleOrder chan int
}

// Select returns the next random backend.
func (b *RandomBackendSelector) Select() string {
	if len(b.Backends) == 0 {
		return ""
	}

	if b.shuffleOrder == nil {
		b.shuffleOrder = make(chan int, len(b.Backends))
		rand.Seed(time.Now().UnixNano())
		for _, v := range rand.Perm(len(b.Backends)) {
			b.shuffleOrder <- v
		}
	}

	next := <-b.shuffleOrder
	b.shuffleOrder <- next

	return b.Backends[next]
}
