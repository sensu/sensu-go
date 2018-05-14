package statser

import (
	"time"

	"github.com/atlassian/gostatsd"
)

// TaggedStatser adds tags and submits metrics to another Statser
type TaggedStatser struct {
	statser Statser
	tags    gostatsd.Tags // invariant: non-empty (otherwise we return the original Statser)
}

// NewTaggedStatser creates a new Statser which adds additional tags
// all metrics submitted.
func NewTaggedStatser(statser Statser, tags gostatsd.Tags) Statser {
	if len(tags) == 0 {
		return statser
	}

	return &TaggedStatser{
		statser: statser,
		tags:    tags,
	}
}

func (ts *TaggedStatser) NotifyFlush(d time.Duration) {
	ts.statser.NotifyFlush(d)
}

func (ts *TaggedStatser) RegisterFlush() (<-chan time.Duration, func()) {
	return ts.statser.RegisterFlush()
}

// Gauge sends a gauge metric
func (ts *TaggedStatser) Gauge(name string, value float64, tags gostatsd.Tags) {
	ts.statser.Gauge(name, value, ts.concatTags(ts.tags, tags))
}

// Count sends a counter metric
func (ts *TaggedStatser) Count(name string, amount float64, tags gostatsd.Tags) {
	ts.statser.Count(name, amount, ts.concatTags(ts.tags, tags))
}

// Increment sends a counter metric with a value of 1
func (ts *TaggedStatser) Increment(name string, tags gostatsd.Tags) {
	ts.statser.Increment(name, ts.concatTags(ts.tags, tags))
}

// TimingMS sends a timing metric from a millisecond value
func (ts *TaggedStatser) TimingMS(name string, ms float64, tags gostatsd.Tags) {
	ts.statser.TimingMS(name, ms, ts.concatTags(ts.tags, tags))
}

// TimingDuration sends a timing metric from a time.Duration
func (ts *TaggedStatser) TimingDuration(name string, d time.Duration, tags gostatsd.Tags) {
	ts.statser.TimingDuration(name, d, ts.concatTags(ts.tags, tags))
}

// NewTimer returns a new timer with time set to now
func (ts *TaggedStatser) NewTimer(name string, tags gostatsd.Tags) *Timer {
	return ts.statser.NewTimer(name, ts.concatTags(ts.tags, tags))
}

// WithTags creates a new Statser with additional tags
func (ts *TaggedStatser) WithTags(tags gostatsd.Tags) Statser {
	// There's no value wrapping it up in multiple layers
	if len(tags) == 0 {
		return ts
	}

	return &TaggedStatser{
		statser: ts.statser, // Base Statser
		tags:    ts.tags.Concat(tags),
	}
}

func (ts *TaggedStatser) concatTags(base, extra gostatsd.Tags) gostatsd.Tags {
	if len(extra) == 0 {
		return base
	}
	return base.Concat(extra)
}
