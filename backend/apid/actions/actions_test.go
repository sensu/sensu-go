// +build !debug

package actions

import (
	"io/ioutil"

	log "github.com/Sirupsen/logrus"
)

func init() {
	// Suppress log output
	log.SetOutput(ioutil.Discard)
}
