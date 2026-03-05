//go:build !debug
// +build !debug

package schedulerd

import (
	"io"

	"github.com/sirupsen/logrus"
)

func init() {
	// Silence logger
	logrus.SetOutput(io.Discard)
}
