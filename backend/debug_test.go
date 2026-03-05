//go:build !debug
// +build !debug

package backend

import (
	"io"

	"github.com/sirupsen/logrus"
)

func init() {
	// Silence logger
	logrus.SetOutput(io.Discard)
}
