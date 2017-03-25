// Package system provides information about the system of the current
// process.
package system

import (
        "testing"

	"github.com/stretchr/testify/assert"
)

func TestInfo(t *testing.T) {
        info, _ := Info()
        assert.NotEmpty(t, info.Hostname)
        assert.NotEmpty(t, info.OS)
        assert.NotEmpty(t, info.Platform)
        assert.NotEmpty(t, info.PlatformFamily)
        assert.NotEmpty(t, info.PlatformVersion)
}
