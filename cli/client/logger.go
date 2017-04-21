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
func (*Logger) Write(p []byte) (n int, err error) {
	n := bytes.IndexByte(byteArray, 0)
	s := string(byteArray[n])

	logger.Debug(s)
}
