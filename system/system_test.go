// Package system provides information about the system of the current
// process.
package system

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInfo(t *testing.T) {
	info, err := Info()
	assert.NoError(t, err)
	assert.NotEmpty(t, info.Arch)
	assert.NotEmpty(t, info.Hostname)
	assert.NotEmpty(t, info.OS)
	assert.NotEmpty(t, info.Platform)
	if info.Platform == "linux" {
		assert.NotEmpty(t, info.PlatformFamily)
	}
	assert.NotEmpty(t, info.PlatformVersion)

}

func TestNetworkInfo(t *testing.T) {
	network, err := NetworkInfo()
	assert.NoError(t, err)
	assert.NotEmpty(t, network.Interfaces)
	nInterface := network.Interfaces[0]
	assert.NotEmpty(t, nInterface.Name)
	// assert.NotEmpty(t, nInterface.MAC) // can be empty
}
