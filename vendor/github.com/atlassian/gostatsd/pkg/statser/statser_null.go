package statser

import (
	"time"

	"github.com/atlassian/gostatsd"
)

// NullStatser is a null implementation of Statser, intended primarily
// for test purposes
type NullStatser struct {
	flushNotifier
}

// NewNullStatser creates a new NullStatser
func NewNullStatser() Statser {
	return &NullStatser{}
}

// Gauge does nothing
func (ns *NullStatser) Gauge(name string, value float64, tags gostatsd.Tags) {}

// Count does nothing
func (ns *NullStatser) Count(name string, amount float64, tags gostatsd.Tags) {}

// Increment does nothing
func (ns *NullStatser) Increment(name string, tags gostatsd.Tags) {}

// TimingMS does nothing
func (ns *NullStatser) TimingMS(name string, ms float64, tags gostatsd.Tags) {}

// TimingDuration does nothing
func (ns *NullStatser) TimingDuration(name string, d time.Duration, tags gostatsd.Tags) {}

// NewTimer returns a new timer with time set to now
func (ns *NullStatser) NewTimer(name string, tags gostatsd.Tags) *Timer {
	return newTimer(ns, name, tags)
}

// WithTags returns a NullStatser
func (ns *NullStatser) WithTags(tags gostatsd.Tags) Statser {
	return ns
}
