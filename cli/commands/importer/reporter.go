package importer

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/sensu/sensu-go/cli/elements/globals"
)

// Reporter reports debug, warning, & errors
type Reporter struct {
	Out      io.Writer
	LogLevel logrus.Level

	entries []*ReporterEntry
}

// Flush pops entries from list and writes them
func (r *Reporter) Flush() error {
	entries := make([]*ReporterEntry, len(r.entries))
	copy(entries, r.entries)
	r.entries = []*ReporterEntry{}

	for _, entry := range entries {
		if entry.Level > r.LogLevel {
			continue
		}

		level := strings.ToUpper(entry.Level.String())
		if entry.Level == logrus.WarnLevel {
			level = globals.WarningStyle(level)
		} else if entry.Level == logrus.ErrorLevel {
			level = globals.ErrorTextStyle(level)
		} else if entry.Level == logrus.InfoLevel {
			level = globals.PrimaryTextStyle(level)
		}

		fmt.Fprintf(
			r.Out,
			"%s\r\t%s\n",
			level,
			entry.Message,
		)
	}

	return nil
}

// Fatal logs fatal error and quits
func (r *Reporter) Fatal(msg string) {
	r.addEntry(logrus.FatalLevel, msg)

	// Flush any remaining entries and exit
	r.Flush()
	os.Exit(1)
}

// Fatalf logs fatal error and quits
func (r *Reporter) Fatalf(msg string, a ...interface{}) {
	r.Fatal(fmt.Sprintf(msg, a...))
}

// Error logs error
func (r *Reporter) Error(msg string) {
	r.addEntry(logrus.ErrorLevel, msg)
}

// Errorf logs error
func (r *Reporter) Errorf(msg string, a ...interface{}) {
	r.Error(fmt.Sprintf(msg, a...))
}

// Warn logs warning
func (r *Reporter) Warn(msg string) {
	r.addEntry(logrus.WarnLevel, msg)
}

// Warnf logs warning
func (r *Reporter) Warnf(msg string, a ...interface{}) {
	r.Warn(fmt.Sprintf(msg, a...))
}

// Info logs information
func (r *Reporter) Info(msg string) {
	r.addEntry(logrus.InfoLevel, msg)
}

// Infof logs info
func (r *Reporter) Infof(msg string, a ...interface{}) {
	r.Info(fmt.Sprintf(msg, a...))
}

// Debug logs debug messages
func (r *Reporter) Debug(msg string) {
	r.addEntry(logrus.DebugLevel, msg)
}

// Debugf logs debug messages
func (r *Reporter) Debugf(msg string, a ...interface{}) {
	r.Debug(fmt.Sprintf(msg, a...))
}

func (r *Reporter) addEntry(level logrus.Level, msg string) *ReporterEntry {
	entry := ReporterEntry{
		Message: msg,
		Level:   level,
		Time:    time.Now(),
	}
	r.entries = append(r.entries, &entry)

	return &entry
}

func (r *Reporter) hasWarnings() bool {
	for _, entry := range r.entries {
		if entry.Level <= logrus.WarnLevel {
			return true
		}
	}

	return false
}

func (r *Reporter) hasErrors() bool {
	for _, entry := range r.entries {
		if entry.Level <= logrus.ErrorLevel {
			return true
		}
	}

	return false
}

// ReporterEntry ...
type ReporterEntry struct {
	Level   logrus.Level
	Time    time.Time
	Message string
	Data    map[string]interface{}
}
