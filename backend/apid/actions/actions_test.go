//go:build !debug
// +build !debug

package actions

import (
	"io/ioutil"

	log "github.com/sirupsen/logrus"
)

func init() {
	// Suppress log output
	log.SetOutput(ioutil.Discard)
}
