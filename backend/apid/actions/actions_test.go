//go:build !debug
// +build !debug

package actions

import (
	"io"

	log "github.com/sirupsen/logrus"
)

func init() {
	// Suppress log output
	log.SetOutput(io.Discard)
}
