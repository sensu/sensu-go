package nodes

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type fakeTime struct {
	samples []time.Time
	next    int
}

func (ft *fakeTime) Now() time.Time {
	t := ft.samples[ft.next]
	ft.next = (ft.next + 1) % len(ft.samples)
	fmt.Printf("%v\n", t)
	return t
}

func TestLearnNode(t *testing.T) {
	t.Parallel()

	now := time.Now()

	rnt := &redisNodeTracker{
		updateInterval: 1 * time.Second,
		expiryInterval: 5 * time.Second,
		now: (&fakeTime{
			samples: []time.Time{
				now.Add(time.Second),
			},
		}).Now,
	}

	// Empty
	node, err := rnt.Select(0)
	assert.NotNil(t, err)
	assert.Equal(t, "", node)

	// Update internal tracking
	rnt.updateNode("127.0.0.1:80")

	// Present
	node, err = rnt.Select(0)
	assert.Nil(t, err)
	assert.Equal(t, "127.0.0.1:80", node)
}

func TestLearnMultipleNodes(t *testing.T) {
	t.Parallel()

	now := time.Now()

	rnt := &redisNodeTracker{
		updateInterval: 1 * time.Second,
		expiryInterval: 5 * time.Second,
		now: (&fakeTime{
			samples: []time.Time{
				now.Add(time.Second),
			},
		}).Now,
	}

	// Empty
	node, err := rnt.Select(0)
	assert.NotNil(t, err)
	assert.Equal(t, "", node)

	// Update internal tracking with multiple nodes (out of sort order)
	rnt.updateNode("127.0.0.2:80")
	rnt.updateNode("127.0.0.1:80")

	nodes := rnt.List()
	assert.Equal(t, nodes, []string{
		"127.0.0.1:80",
		"127.0.0.2:80",
	})
}

func TestForgetNodes(t *testing.T) {
	t.Parallel()

	now := time.Now()

	rnt := &redisNodeTracker{
		updateInterval: 1 * time.Second,
		expiryInterval: 5 * time.Second,
		now: (&fakeTime{
			samples: []time.Time{
				now,
				now.Add(1000 * time.Millisecond),
				now.Add(5500 * time.Millisecond),
			},
		}).Now,
	}

	// Empty
	node, err := rnt.Select(0)
	assert.NotNil(t, err)
	assert.Equal(t, "", node)

	// Update internal tracking with multiple nodes
	rnt.updateNode("127.0.0.1:80") // time.sample[0] - will expire
	rnt.updateNode("127.0.0.2:80") // time.sample[1] - will not expire
	rnt.expireNodes()              // time.sample[2]

	node, err = rnt.Select(0)
	assert.Nil(t, err)
	assert.Equal(t, "127.0.0.2:80", node)

	node, err = rnt.Select(1)
	assert.Nil(t, err)
	assert.Equal(t, "127.0.0.2:80", node)
}

func TestUpdateExistingNode(t *testing.T) {
	t.Parallel()

	now := time.Now()

	rnt := &redisNodeTracker{
		updateInterval: 1 * time.Second,
		expiryInterval: 5 * time.Second,
		now: (&fakeTime{
			samples: []time.Time{
				now,
				now.Add(1000 * time.Millisecond),
				now.Add(5500 * time.Millisecond),
			},
		}).Now,
	}

	// Empty
	node, err := rnt.Select(0)
	assert.NotNil(t, err)
	assert.Equal(t, "", node)

	// Update internal tracking with multiple nodes
	rnt.updateNode("127.0.0.1:80") // time.sample[0] - now
	rnt.updateNode("127.0.0.1:80") // time.sample[1] - now+1 - refreshes timestamp
	rnt.expireNodes()              // time.sample[2] - now+5.5

	// Make sure only one node
	nodes := rnt.List()
	assert.Equal(t, nodes, []string{
		"127.0.0.1:80",
	})
}

var s string
var e error
var ss []string

func BenchmarkSelects(b *testing.B) {
	rnt := &redisNodeTracker{
		now: time.Now,
	}

	rnt.updateNode("127.0.0.1:80")
	rnt.updateNode("127.0.0.2:80")

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		s, e = rnt.Select(uint64(n))
	}
}

func BenchmarkLists(b *testing.B) {
	rnt := &redisNodeTracker{
		now: time.Now,
	}

	rnt.updateNode("127.0.0.1:80")
	rnt.updateNode("127.0.0.2:80")

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		ss = rnt.List()
	}
}

func BenchmarkParallelSelects(b *testing.B) {
	rnt := &redisNodeTracker{
		now: time.Now,
	}

	rnt.updateNode("127.0.0.1:80")
	rnt.updateNode("127.0.0.2:80")

	// Create contention on the lock
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				rnt.updateNode("127.0.0.1:80")
				rnt.updateNode("127.0.0.2:80")
			}
		}
	}()

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		s, e = rnt.Select(uint64(n))
	}
	b.StopTimer()

	cancel()
	wg.Wait()
}
