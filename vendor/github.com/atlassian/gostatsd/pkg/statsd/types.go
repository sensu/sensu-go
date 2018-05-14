package statsd

import (
	"context"
	"time"

	"github.com/atlassian/gostatsd"
	"github.com/atlassian/gostatsd/pkg/statser"
)

// MetricHandler can be used to handle metrics
type MetricHandler interface {
	// EstimatedTags returns a guess for how many tags to pre-allocate
	EstimatedTags() int
	// DispatchMetric dispatches a metric to the next step in a pipeline.
	DispatchMetric(ctx context.Context, m *gostatsd.Metric) error
}

// EventHandler can be used to handle events
type EventHandler interface {
	// DispatchEvent dispatches event to the next step in a pipeline.
	DispatchEvent(ctx context.Context, e *gostatsd.Event) error
	// WaitForEvents waits for all event-dispatching goroutines to finish.
	WaitForEvents()
}

// DispatcherProcessFunc is a function that gets executed by Dispatcher for each Aggregator, passing it into the function.
type DispatcherProcessFunc func(int, Aggregator)

// AggregateProcesser is an interface to run a function against each Aggregator, in the goroutine
// context of that Aggregator.
type AggregateProcesser interface {
	Process(ctx context.Context, fn DispatcherProcessFunc) gostatsd.Wait
}

// ProcessFunc is a function that gets executed by Aggregator with its state passed into the function.
type ProcessFunc func(*gostatsd.MetricMap)

// Aggregator is an object that aggregates statsd metrics.
// The function NewAggregator should be used to create the objects.
//
// Incoming metrics should be passed via Receive function.
type Aggregator interface {
	Receive(*gostatsd.Metric, time.Time)
	Flush(interval time.Duration)
	Process(ProcessFunc)
	Reset()
}

// Datagram is a received UDP datagram that has not been parsed into Metric/Event(s)
type Datagram struct {
	IP       gostatsd.IP
	Msg      []byte
	DoneFunc func() // to be called once the datagram has been parsed and msg can be freed
}

// MetricEmitter is an object that emits metrics.  Used to pass a Statser to the object
// after initialization, as Statsers may be created after MetricEmitters
type MetricEmitter interface {
	RunMetrics(ctx context.Context, statser statser.Statser)
}
