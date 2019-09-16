package agentd

import "github.com/sirupsen/logrus"

var logger = logrus.WithFields(logrus.Fields{
	"component": "agentd",
})

type logrusIOWriter struct {
	entry *logrus.Entry
}

// Write satifies the io.Writer interface
func (w *logrusIOWriter) Write(b []byte) (int, error) {
	n := len(b)

	// Remove newline
	if n > 0 && b[n-1] == '\n' {
		b = b[:n-1]
	}

	w.entry.Warning(string(b))
	return n, nil
}
