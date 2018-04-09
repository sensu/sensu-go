package report

import (
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

// Writer exposes methods for adding entries to a report
type Writer struct {
	report  *Report
	context map[string]interface{}
}

// NewWriter returns a new writer instance with context initialized
func NewWriter(report *Report) Writer {
	return Writer{
		report:  report,
		context: make(map[string]interface{}),
	}
}

// WithValue returns a new writer instance with KV pair added
func (w Writer) WithValue(key string, val interface{}) Writer {
	newContext := make(map[string]interface{}, len(w.context)+1)
	for k, v := range w.context {
		newContext[k] = v
	}
	newContext[key] = val

	return Writer{report: w.report, context: newContext}
}

// Fatal logs fatal error and quits
func (w Writer) Fatal(msg string) {
	w.addEntry(logrus.FatalLevel, msg)

	// Flush any remaining entries and exit
	_ = w.report.Flush()
	os.Exit(1)
}

// Fatalf logs fatal error and quits
func (w Writer) Fatalf(msg string, a ...interface{}) {
	w.Fatal(fmt.Sprintf(msg, a...))
}

// Error logs error
func (w Writer) Error(msg string) {
	w.addEntry(logrus.ErrorLevel, msg)
}

// Errorf logs error
func (w Writer) Errorf(msg string, a ...interface{}) {
	w.Error(fmt.Sprintf(msg, a...))
}

// Warn logs warning
func (w Writer) Warn(msg string) {
	w.addEntry(logrus.WarnLevel, msg)
}

// Warnf logs warning
func (w Writer) Warnf(msg string, a ...interface{}) {
	w.Warn(fmt.Sprintf(msg, a...))
}

// Info logs information
func (w Writer) Info(msg string) {
	w.addEntry(logrus.InfoLevel, msg)
}

// Infof logs info
func (w Writer) Infof(msg string, a ...interface{}) {
	w.Info(fmt.Sprintf(msg, a...))
}

// Debug logs debug messages
func (w Writer) Debug(msg string) {
	w.addEntry(logrus.DebugLevel, msg)
}

// Debugf logs debug messages
func (w Writer) Debugf(msg string, a ...interface{}) {
	w.Debug(fmt.Sprintf(msg, a...))
}

func (w *Writer) addEntry(level logrus.Level, msg string) {
	entry := Entry{
		Message: msg,
		Context: w.context,
		Level:   level,
		Time:    time.Now(),
	}
	w.report.AddEntry(entry)
}
