package statser

import (
	"time"

	"github.com/atlassian/gostatsd"
)

// Timer times an operation and submits a timing or gauge metric
type Timer struct {
	statser   Statser
	name      string
	tags      gostatsd.Tags
	startTime time.Time
	endTime   time.Time
}

func newTimer(statser Statser, name string, tags gostatsd.Tags) *Timer {
	return &Timer{
		statser:   statser,
		name:      name,
		tags:      tags,
		startTime: time.Now(),
	}
}

// Stop stops the time being recorded
func (t *Timer) Stop() {
	t.endTime = time.Now()
}

// Send sends a timing style metric
func (t *Timer) Send() {
	t.statser.TimingDuration(t.name, t.duration(), t.tags)
}

// SendGauge sends a gauge metric in milliseconds
func (t *Timer) SendGauge() {
	t.statser.Gauge(t.name, float64(t.duration())/float64(time.Millisecond), t.tags)
}

func (t *Timer) duration() time.Duration {
	if t.endTime.IsZero() {
		return time.Since(t.startTime)
	}
	return t.endTime.Sub(t.startTime)
}
