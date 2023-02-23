package testspec

import (
	"context"
	"errors"
	"math/rand"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/sensu/sensu-go/backend/queue"
)

func RunClientTestSuite(t *testing.T, ctx context.Context, clientUnderTest queue.Client) {
	t.Run("integration test suite", func(t *testing.T) {
		runIntegrationSuite(t, ctx, clientUnderTest)
	})
	t.Run("queue isolation", func(t *testing.T) {
		runQueueIsolationSuite(t, ctx, clientUnderTest)
	})
	t.Run("reservation timeout", func(t *testing.T) {
		runReservationTimeout(t, ctx, clientUnderTest)
	})
}

// runIntegrationSuite attempts to simulate heavy utilization of a queue.
// It feeds the queue with a set of items, starts workers that either Ack
// or Nack their reservation after a some random interval, and validates that
// all published items were recieved exactly once.
func runIntegrationSuite(t *testing.T, ctx context.Context, clientUnderTest queue.Client) {
	var toEnqueue []queue.Item
	for i := uint8(0); i < 100; i++ {
		item := queue.Item{
			Queue: "test-queue",
			Value: []byte{i},
		}
		toEnqueue = append(toEnqueue, item)
	}

	enqueueFn := func(q queue.Client, items []queue.Item) {
		for _, item := range items {
			q.Enqueue(ctx, item)
			jitter := time.Duration(20 * rand.Float64())
			time.Sleep(time.Millisecond * jitter)
		}
	}

	var wg sync.WaitGroup

	actualResults := make([]queue.Item, 100)

	nackerFn := func(q queue.Client) {
		defer wg.Done()
		// Do not poll indefinately
		rCtx, cancel := context.WithTimeout(ctx, time.Millisecond*50)
		defer cancel()

		res, err := q.Reserve(rCtx, "test-queue")
		if err != nil {
			if errors.Is(err, rCtx.Err()) {
				return
			}
			t.Errorf("unexpected reservation error: %v", err)
			return
		}
		jitter := time.Duration(20 * rand.Float64())
		time.Sleep(time.Millisecond * jitter)

		// return item to queue
		res.Nack(ctx)

	}
	workerFn := func(q queue.Client, n int) {
		defer wg.Done()
		res, err := q.Reserve(ctx, "test-queue")
		if err != nil {
			t.Errorf("unexpected reservation error: %v", err)
		}
		jitter := time.Duration(20 * rand.Float64())
		time.Sleep(time.Millisecond * jitter)

		actualResults[n] = res.Item()
		if err := res.Ack(ctx); err != nil {
			t.Errorf("unexpected ack error: %v", err)
		}
	}

	go enqueueFn(clientUnderTest, toEnqueue)
	wg.Add(20)
	// Reserve but then Nack 20 items
	for i := 0; i < 20; i++ {
		go nackerFn(clientUnderTest)
	}
	wg.Add(100)
	// Reserve and Ack 100, recording the restults
	for i := 0; i < 100; i++ {
		go workerFn(clientUnderTest, i)
	}

	wg.Wait()

	// NOT a FIFO, order delivery not guaranteed
	sort.Slice(actualResults, func(i, j int) bool {
		return actualResults[i].Value[0] < actualResults[j].Value[0]
	})

	for idx, result := range actualResults {
		got := result.Value[0]
		want := uint8(idx)
		if got != want {
			t.Errorf("unexpected result order. wanted: %d got: %d - id: %s", want, got, result.ID)
		}
	}
}

// runQueueIsolationSuite publishes to multiple queues and ensures delivery
// happens on the appropriate queue.
func runQueueIsolationSuite(t *testing.T, ctx context.Context, clientUnderTest queue.Client) {

	err := clientUnderTest.Enqueue(ctx, queue.Item{
		Queue: "queue-1",
		Value: []byte("hello world")},
	)
	if err != nil {
		t.Errorf("unexpected enqueue error: %v", err)
		return
	}
	done := make(chan struct{})
	go func() {
		clientUnderTest.Reserve(ctx, "queue-2")
		close(done)
	}()

	select {
	case <-time.After(time.Millisecond * 100):
		// okay
	case <-done:
		t.Error("unexpected reservation on queue-2")
	}
	resQueue1, _ := clientUnderTest.Reserve(ctx, "queue-1")
	resQueue1.Ack(ctx)

	clientUnderTest.Enqueue(ctx, queue.Item{
		Queue: "queue-2",
		Value: []byte("hello world")},
	)

	select {
	case <-time.After(time.Second):
		t.Error("expected reservation on queue-2")
	case <-done:
		// okay
	}

}

// runReservationTimeout checks that a blocked Reserve call times out when the
// context expires.
func runReservationTimeout(t *testing.T, ctx context.Context, clientUnderTest queue.Client) {
	tCtx, cancel := context.WithTimeout(ctx, time.Millisecond*100)
	defer cancel()

	reserveErr := make(chan error)
	go func() {
		_, err := clientUnderTest.Reserve(tCtx, "empty-queue")
		reserveErr <- err
		close(reserveErr)
	}()
	select {
	case <-time.After(time.Millisecond * 500):
		t.Errorf("expected queue Reservaiton to time out")
	case err := <-reserveErr:
		if got, want := err, tCtx.Err(); got != want {
			t.Errorf("unexpected error reserving queue item: got %v, want %v", got, want)
		}
	}
}
