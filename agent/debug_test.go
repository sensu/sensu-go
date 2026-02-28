//go:build !debug
// +build !debug

package agent

import (
	"io"

	"github.com/sirupsen/logrus"
)

func init() {
	// Silence logger
	logrus.SetOutput(io.Discard)
}
