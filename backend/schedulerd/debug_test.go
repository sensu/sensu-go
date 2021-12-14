//go:build !debug
// +build !debug

package schedulerd

import (
	"io/ioutil"

	"github.com/sirupsen/logrus"
)

func init() {
	// Silence logger
	logrus.SetOutput(ioutil.Discard)
}
