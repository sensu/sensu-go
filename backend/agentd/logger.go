package agentd

import (
	"bytes"
	"github.com/sensu/sensu-go/util/logging"

	"github.com/sirupsen/logrus"
)

var logger = logging.GetLogger("agentd").WithFields(logrus.Fields{
	"component": "agentd",
})

type logrusIOWriter struct {
	entry *logrus.Entry
}

// Write satifies the io.Writer interface
func (w *logrusIOWriter) Write(b []byte) (int, error) {
	n := len(b)

	// Remove all leading and trailing white space
	b = bytes.TrimSpace(b)

	w.entry.Warning(string(b))
	return n, nil
}
