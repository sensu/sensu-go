// +build !debug

package etcd

import (
	"io/ioutil"

	"github.com/Sirupsen/logrus"
)

func init() {
	// Silence logger
	logrus.SetOutput(ioutil.Discard)
}
