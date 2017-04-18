package etcd

import (
	"fmt"

	log "github.com/Sirupsen/logrus"

	"github.com/coreos/pkg/capnslog"
)

type logrusFormatter struct {
	logger *log.Entry
}

func NewLogrusFormatter() capnslog.Formatter {
	logger := log.WithFields(log.Fields{
		"component": "etcd",
	})

	return &logrusFormatter{
		logger: logger,
	}
}

func (s *logrusFormatter) logWithPkg(pkg string) *log.Entry {
	return s.logger.WithFields(log.Fields{
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
