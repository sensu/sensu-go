package etcd

import (
	"fmt"

	log "github.com/Sirupsen/logrus"

	"github.com/coreos/pkg/capnslog"
)

func NewLogrusFormatter() capnslog.Formatter {
	return &logrusFormatter{}
}

func logWithPkg(pkg string) *log.Entry {
	return log.WithFields(log.Fields{
		"component": "etcd",
		"pkg":       pkg,
	})
}

type logrusFormatter struct{}

// Format builds a log message for the LogrusFormatter.
func (s *logrusFormatter) Format(pkg string, l capnslog.LogLevel, _ int, entries ...interface{}) {
	for _, entry := range entries {
		str := fmt.Sprint(entry)
		switch l {
		case capnslog.CRITICAL:
			logWithPkg(pkg).Fatal(str)
		case capnslog.ERROR:
			logWithPkg(pkg).Error(str)
		case capnslog.WARNING:
			logWithPkg(pkg).Warning(str)
		case capnslog.NOTICE:
			logWithPkg(pkg).Warning(str)
		case capnslog.INFO:
			logWithPkg(pkg).Info(str)
		case capnslog.DEBUG:
			logWithPkg(pkg).Debug(str)
		case capnslog.TRACE:
			logWithPkg(pkg).Debug(str)
		default:
			panic("Unhandled loglevel")
		}
	}
}

// Flush is included so that the interface is complete, but is a no-op.
func (s *logrusFormatter) Flush() {
}
