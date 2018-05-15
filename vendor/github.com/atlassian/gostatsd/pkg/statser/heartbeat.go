package statser

import (
	"context"

	"github.com/atlassian/gostatsd"
)

// HeartBeater periodically sends a gauge for heartbeat purposes
type HeartBeater struct {
	statser    Statser
	metricName string
}

// NewHeartBeater creates a new HeartBeater
func NewHeartBeater(statser Statser, metricName string, tags gostatsd.Tags) *HeartBeater {
	return &HeartBeater{
		statser:    statser.WithTags(tags),
		metricName: metricName,
	}
}

// Run will run a HeartBeater in the background until the supplied context is closed.
func (hb *HeartBeater) Run(ctx context.Context) {
	flushed, unregister := hb.statser.RegisterFlush()
	defer unregister()

	for {
		select {
		case <-ctx.Done():
			return
		case <-flushed:
			hb.emit()
		}
	}
}

func (hb *HeartBeater) emit() {
	hb.statser.Gauge(hb.metricName, 1, nil)
}
