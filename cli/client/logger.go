package client

import (
	"bytes"

	"github.com/Sirupsen/logrus"
)

var logger = logrus.WithFields(logrus.Fields{
	"component": "cli-client",
})

// Logger ...
type Logger struct{}

// Write implements io.Writer interface
func (*Logger) Write(p []byte) (int, error) {
	n := bytes.IndexByte(p, 0)
	s := string(p[n])

	logger.Debug(s)

	return 0, nil
}
