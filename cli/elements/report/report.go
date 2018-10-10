package report

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/sensu/sensu-go/cli/elements/globals"
)

// Report reports debug, warning, & errors
type Report struct {
	Out      io.Writer
	LogLevel logrus.Level

	entries []Entry
}

// New returns new report w/ log level preconfigured
func New() Report {
	return Report{LogLevel: logrus.InfoLevel}
}

// Flush pops entries from list and writes them
func (r *Report) Flush() error {
	entries := make([]Entry, len(r.entries))
	copy(entries, r.entries)
	r.entries = []Entry{}

	for _, entry := range entries {
		if entry.Level > r.LogLevel {
			continue
		}

		level := strings.ToUpper(entry.Level.String())
		if entry.Level == logrus.WarnLevel {
			level = globals.WarningStyle(level)
		} else if entry.Level <= logrus.ErrorLevel {
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

// HasWarnings returns true if there are any warnings entries in the report
func (r *Report) HasWarnings() bool {
	for _, entry := range r.entries {
		if entry.Level == logrus.WarnLevel {
			return true
		}
	}

	return false
}

// HasErrors returns true if there are any error entries in the report
func (r *Report) HasErrors() bool {
	for _, entry := range r.entries {
		if entry.Level <= logrus.ErrorLevel {
			return true
		}
	}

	return false
}

// AddEntry adds given entry to report
func (r *Report) AddEntry(e Entry) {
	r.entries = append(r.entries, e)
}

// Entry ...
type Entry struct {
	Level   logrus.Level
	Message string
	Context map[string]interface{}
	Time    time.Time
}
