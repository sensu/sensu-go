//go:build !debug
// +build !debug

package etcd

import (
	io "io"

	"github.com/sirupsen/logrus"
)

func init() {
	// Silence logger
	logrus.SetOutput(io.Discard)
}
