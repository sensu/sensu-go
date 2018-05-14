package statser

import (
	"time"

	"github.com/atlassian/gostatsd"

	log "github.com/sirupsen/logrus"
)

// LoggingStatser is a Statser which emits logs
type LoggingStatser struct {
	flushNotifier

	tags   gostatsd.Tags
	logger *log.Entry
}

// NewLoggingStatser creates a new Statser which sends metrics to the
// supplied log.Entry
func NewLoggingStatser(tags gostatsd.Tags, logger *log.Entry) Statser {
	return &LoggingStatser{
		tags:   tags,
		logger: logger,
	}
}

// Gauge sends a gauge metric
func (ls *LoggingStatser) Gauge(name string, value float64, tags gostatsd.Tags) {
	ls.logger.WithFields(log.Fields{
		"name":  name,
		"tags":  ls.tags.Concat(tags),
		"value": value,
	}).Infof("gauge")
}

// Count sends a counter metric
func (ls *LoggingStatser) Count(name string, amount float64, tags gostatsd.Tags) {
	ls.logger.WithFields(log.Fields{
		"name":   name,
		"tags":   ls.tags.Concat(tags),
		"amount": amount,
	}).Infof("count")
}

// Increment sends a counter metric with a value of 1
func (ls *LoggingStatser) Increment(name string, tags gostatsd.Tags) {
	ls.logger.WithFields(log.Fields{
		"name": name,
		"tags": ls.tags.Concat(tags),
	}).Infof("increment")
}

// TimingMS sends a timing metric from a millisecond value
func (ls *LoggingStatser) TimingMS(name string, ms float64, tags gostatsd.Tags) {
	ls.logger.WithFields(log.Fields{
		"name": name,
		"tags": ls.tags.Concat(tags),
		"ms":   ms,
	}).Infof("timing")
}

// TimingDuration sends a timing metric from a time.Duration
func (ls *LoggingStatser) TimingDuration(name string, d time.Duration, tags gostatsd.Tags) {
	ls.TimingMS(name, float64(d)/float64(time.Millisecond), tags)
}

// NewTimer returns a new timer with time set to now
func (ls *LoggingStatser) NewTimer(name string, tags gostatsd.Tags) *Timer {
	return newTimer(ls, name, tags)
}

// WithTags creates a new Statser with additional tags
func (ls *LoggingStatser) WithTags(tags gostatsd.Tags) Statser {
	return NewTaggedStatser(ls, tags)
}
