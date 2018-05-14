package statser

import (
	"time"

	"github.com/atlassian/gostatsd"
)

// Statser is the interface for sending metrics
type Statser interface {
	// NotifyFlush is called when a flush occurs.  It signals all known subscribers.
	NotifyFlush(d time.Duration)
	// RegisterFlush returns a channel which will receive a notification after every flush, and a cleanup
	// function which should be called to signal the channel is no longer being monitored.  If the channel
	// blocks, the notification will be silently dropped.
	RegisterFlush() (ch <-chan time.Duration, unregister func())

	Gauge(name string, value float64, tags gostatsd.Tags)
	Count(name string, amount float64, tags gostatsd.Tags)
	Increment(name string, tags gostatsd.Tags)
	TimingMS(name string, ms float64, tags gostatsd.Tags)
	TimingDuration(name string, d time.Duration, tags gostatsd.Tags)
	NewTimer(name string, tags gostatsd.Tags) *Timer
	WithTags(tags gostatsd.Tags) Statser
}
