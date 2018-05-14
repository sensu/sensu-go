package statser

import (
	"context"
	"time"

	"github.com/atlassian/gostatsd"
)

// ChannelStatsWatcher reports metrics about channel usage to a Statser
type ChannelStatsWatcher struct {
	samples     int
	accumulated int
	min         int
	max         int
	last        int
	capacity    int

	statser        Statser
	lenFunc        func() int
	sampleInterval time.Duration
}

// NewChannelStatsWatcher creates a new ChannelStatsWatcher
func NewChannelStatsWatcher(statser Statser, channelName string, tags gostatsd.Tags, capacity int, lenFunc func() int, sampleInterval time.Duration) *ChannelStatsWatcher {
	return &ChannelStatsWatcher{
		statser:        statser.WithTags(tags.Concat(gostatsd.Tags{"channel:" + channelName})),
		capacity:       capacity,
		lenFunc:        lenFunc,
		sampleInterval: sampleInterval,
	}
}

// Run will run a ChannelStatsWatcher in the background until the supplied context is
// closed.  Metrics are sampled every sampleInterval, and written every flush.
func (csw *ChannelStatsWatcher) Run(ctx context.Context) {
	flushed, unregister := csw.statser.RegisterFlush()
	defer unregister()

	ticker := time.NewTicker(csw.sampleInterval)
	defer ticker.Stop()

	csw.sample()

	for {
		select {
		case <-ctx.Done():
			return
		case <-flushed:
			csw.emit()
			csw.sample() // Ensure there will always be at least one sample
		case <-ticker.C:
			csw.sample()
		}
	}
}

func (csw *ChannelStatsWatcher) sample() {
	csw.samples++
	s := csw.lenFunc()
	csw.accumulated += s
	if s < csw.min {
		csw.min = s
	}
	if s > csw.max {
		csw.max = s
	}
	csw.last = s
}

func (csw *ChannelStatsWatcher) emit() {
	samples := float64(csw.samples)
	csw.statser.Gauge("channel.samples", samples, nil)
	csw.statser.Gauge("channel.avg", float64(csw.accumulated)/samples, nil)
	csw.statser.Gauge("channel.min", float64(csw.min), nil)
	csw.statser.Gauge("channel.max", float64(csw.max), nil)
	csw.statser.Gauge("channel.last", float64(csw.last), nil)
	csw.statser.Gauge("channel.capacity", float64(csw.capacity), nil)

	csw.samples = 0
	csw.accumulated = 0
	csw.min = csw.capacity
	csw.max = 0
}
