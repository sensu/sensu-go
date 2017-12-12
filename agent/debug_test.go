// +build !debug

package agent

import (
	"io/ioutil"

	"github.com/Sirupsen/logrus"
)

func init() {
	// Silence logger
	logrus.SetOutput(ioutil.Discard)
}
