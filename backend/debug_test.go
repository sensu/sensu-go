// +build !debug

package backend

import (
	"io/ioutil"

	"github.com/Sirupsen/logrus"
)

func init() {
	// Silence logger
	logrus.SetOutput(ioutil.Discard)
}
