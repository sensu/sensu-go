//go:build !debug
// +build !debug

package backend

import (
	"io/ioutil"

	"github.com/sirupsen/logrus"
)

func init() {
	// Silence logger
	logrus.SetOutput(ioutil.Discard)
}
