package statsd

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/atlassian/gostatsd"
	"github.com/atlassian/gostatsd/pkg/statser"

	"github.com/ash2k/stager/wait"
	"github.com/stretchr/testify/assert"
)

type testAggregator struct {
	agrNumber int
	af        *testAggregatorFactory
	gostatsd.MetricMap
}

func (a *testAggregator) TrackMetrics(statser statser.Statser) {
}

func (a *testAggregator) Receive(m *gostatsd.Metric, t time.Time) {
	a.af.Mutex.Lock()
	defer a.af.Mutex.Unlock()
	a.af.receiveInvocations[a.agrNumber]++
}

func (a *testAggregator) Flush(interval time.Duration) {
	a.af.Mutex.Lock()
	defer a.af.Mutex.Unlock()
	a.af.flushInvocations[a.agrNumber]++
}

func (a *testAggregator) Process(f ProcessFunc) {
	a.af.Mutex.Lock()
	defer a.af.Mutex.Unlock()
	a.af.processInvocations[a.agrNumber]++
	f(&a.MetricMap)
}

func (a *testAggregator) Reset() {
	a.af.Mutex.Lock()
	defer a.af.Mutex.Unlock()
	a.af.resetInvocations[a.agrNumber]++
}

type testAggregatorFactory struct {
	sync.Mutex
	receiveInvocations map[int]int
	flushInvocations   map[int]int
	processInvocations map[int]int
	resetInvocations   map[int]int
	numAgrs            int
}

func (af *testAggregatorFactory) Create() Aggregator {
	agrNumber := af.numAgrs
	af.numAgrs++
	af.receiveInvocations[agrNumber] = 0
	af.flushInvocations[agrNumber] = 0
	af.processInvocations[agrNumber] = 0
	af.resetInvocations[agrNumber] = 0

	agr := testAggregator{
		agrNumber: agrNumber,
		af:        af,
	}
	agr.Counters = gostatsd.Counters{}
	agr.Timers = gostatsd.Timers{}
	agr.Gauges = gostatsd.Gauges{}
	agr.Sets = gostatsd.Sets{}

	return &agr
}

func newTestFactory() *testAggregatorFactory {
	return &testAggregatorFactory{
		receiveInvocations: make(map[int]int),
		flushInvocations:   make(map[int]int),
		processInvocations: make(map[int]int),
		resetInvocations:   make(map[int]int),
	}
}

func TestNewBackendHandlerShouldCreateCorrectNumberOfWorkers(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	n := r.Intn(5) + 1
	factory := newTestFactory()
	h := NewBackendHandler(nil, 0, n, 1, factory)
	assert.Equal(t, n, len(h.workers))
	assert.Equal(t, n, factory.numAgrs)
}

func TestRunShouldReturnWhenContextCancelled(t *testing.T) {
	t.Parallel()
	h := NewBackendHandler(nil, 0, 5, 1, newTestFactory())
	ctx, cancelFunc := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancelFunc()
	h.Run(ctx)
}

func TestDispatchMetricShouldDistributeMetrics(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	n := r.Intn(5) + 1
	factory := newTestFactory()
	h := NewBackendHandler(nil, 0, n, 10, factory)
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	var wgFinish wait.Group
	wgFinish.StartWithContext(ctx, h.Run)
	numMetrics := r.Intn(1000) + n*10
	var wg sync.WaitGroup
	wg.Add(numMetrics)
	for i := 0; i < numMetrics; i++ {
		m := &gostatsd.Metric{
			Type:  gostatsd.COUNTER,
			Name:  fmt.Sprintf("counter.metric.%d", r.Int63()),
			Tags:  nil,
			Value: r.Float64(),
		}
		go func() {
			defer wg.Done()
			assert.NoError(t, h.DispatchMetric(ctx, m))
		}()
	}
	wg.Wait()       // Wait for all metrics to be dispatched
	cancelFunc()    // After all metrics have been dispatched, we signal dispatcher to shut down
	wgFinish.Wait() // Wait for dispatcher to shutdown

	receiveInvocations := getTotalInvocations(factory.receiveInvocations)
	assert.Equal(t, numMetrics, receiveInvocations)
	for agrNum, count := range factory.receiveInvocations {
		if count == 0 {
			t.Errorf("aggregator %d was never invoked", agrNum)
		} else {
			t.Logf("aggregator %d was invoked %d time(s)", agrNum, count)
		}
	}
}

func getTotalInvocations(inv map[int]int) int {
	var counter int
	for _, i := range inv {
		counter += i
	}
	return counter
}

func BenchmarkBackendHandler(b *testing.B) {
	rand.Seed(time.Now().UnixNano())
	factory := newTestFactory()
	h := NewBackendHandler(nil, 0, runtime.NumCPU(), 10, factory)
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	var wgFinish wait.Group
	wgFinish.StartWithContext(ctx, h.Run)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m := &gostatsd.Metric{
				Type:  gostatsd.COUNTER,
				Name:  fmt.Sprintf("counter.metric.%d", rand.Int63()),
				Tags:  nil,
				Value: rand.Float64(),
			}
			if err := h.DispatchMetric(ctx, m); err != nil {
				b.Errorf("unexpected error: %v", err)
			}
		}
	})
	cancelFunc()    // After all metrics have been dispatched, we signal dispatcher to shut down
	wgFinish.Wait() // Wait for dispatcher to shutdown
}
