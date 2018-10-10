package etcd

import (
	"fmt"

	"github.com/coreos/pkg/capnslog"
	"github.com/sirupsen/logrus"
)

type logrusFormatter struct {
	logger *logrus.Entry
}

// NewLogrusFormatter creates a new LogrusFormatter
func NewLogrusFormatter() capnslog.Formatter {
	logger := logrus.WithFields(logrus.Fields{
		"component": "etcd",
	})

	return &logrusFormatter{
		logger: logger,
	}
}

func (s *logrusFormatter) logWithPkg(pkg string) *logrus.Entry {
	return s.logger.WithFields(logrus.Fields{
		"pkg": pkg,
	})
}

// Format builds a log message for the LogrusFormatter.
func (s *logrusFormatter) Format(pkg string, l capnslog.LogLevel, _ int, entries ...interface{}) {
	for _, entry := range entries {
		str := fmt.Sprint(entry)
		switch l {
		case capnslog.CRITICAL:
			s.logWithPkg(pkg).Fatal(str)
		case capnslog.ERROR:
			s.logWithPkg(pkg).Error(str)
		case capnslog.WARNING:
			s.logWithPkg(pkg).Warning(str)
		case capnslog.NOTICE:
			s.logWithPkg(pkg).Warning(str)
		case capnslog.INFO:
			s.logWithPkg(pkg).Info(str)
		case capnslog.DEBUG:
			s.logWithPkg(pkg).Debug(str)
		case capnslog.TRACE:
			s.logWithPkg(pkg).Debug(str)
		default:
			panic("Unhandled loglevel")
		}
	}
}

// Flush is included so that the interface is complete, but is a no-op.
func (s *logrusFormatter) Flush() {
}
