// Package system provides information about the system of the current
// process.
package system

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInfo(t *testing.T) {
	info, _ := Info()
	assert.NotEmpty(t, info.Arch)
	assert.NotEmpty(t, info.Hostname)
	assert.NotEmpty(t, info.OS)
	assert.NotEmpty(t, info.Platform)
	if info.Platform == "linux" {
		assert.NotEmpty(t, info.PlatformFamily)
	}
	assert.NotEmpty(t, info.PlatformVersion)
	assert.NotEmpty(t, info.Network.Interfaces)
	nInterface := info.Network.Interfaces[0]
	assert.NotEmpty(t, nInterface.Name)
	//assert.NotEmpty(t, nInterface.MAC) // can be empty
	assert.NotEmpty(t, nInterface.Addresses)
}
