//go:build !debug
// +build !debug

package etcd

import (
	"io/ioutil"

	"github.com/sirupsen/logrus"
)

func init() {
	// Silence logger
	logrus.SetOutput(ioutil.Discard)
}
