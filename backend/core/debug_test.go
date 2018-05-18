// +build !debug

package core

import (
	"io/ioutil"

	"github.com/sirupsen/logrus"
)

func init() {
	// Silence logger
	logrus.SetOutput(ioutil.Discard)
}
