// Package system provides information about the system of the current
// process.
package system

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInfo(t *testing.T) {
	// Test first with network information included
	info, err := Info(true)
	assert.NoError(t, err)
	assert.NotEmpty(t, info.Arch)
	assert.NotEmpty(t, info.Hostname)
	assert.NotEmpty(t, info.OS)
	assert.NotEmpty(t, info.Platform)
	if info.Platform == "linux" {
		assert.NotEmpty(t, info.PlatformFamily)
	}
	assert.NotEmpty(t, info.PlatformVersion)
	assert.NoError(t, err)
	assert.NotEmpty(t, info.Network.Interfaces)
	nInterface := info.Network.Interfaces[0]
	assert.NotEmpty(t, nInterface.Name)

	// Then we have to test if with network data stripped
	infoWithoutNet, err := Info(false)
	assert.NoError(t, err)
	assert.NotEmpty(t, infoWithoutNet.Arch)
	assert.NotEmpty(t, infoWithoutNet.Hostname)
	assert.NotEmpty(t, infoWithoutNet.OS)
	assert.NotEmpty(t, infoWithoutNet.Platform)
	if info.Platform == "linux" {
		assert.NotEmpty(t, infoWithoutNet.PlatformFamily)
	}
	assert.NotEmpty(t, infoWithoutNet.PlatformVersion)
	assert.NoError(t, err)
	assert.Empty(t, infoWithoutNet.Network.Interfaces)
}

func TestNetworkInfo(t *testing.T) {
	network, err := NetworkInfo()
	assert.NoError(t, err)
	assert.NotEmpty(t, network.Interfaces)
	nInterface := network.Interfaces[0]
	assert.NotEmpty(t, nInterface.Name)
	// assert.NotEmpty(t, nInterface.MAC) // can be empty
}
