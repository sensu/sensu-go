package actions

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewVersionController(t *testing.T) {
	assert := assert.New(t)
	actions := NewVersionController("0.0.0")

	assert.NotNil(actions)
	assert.Equal(actions.clusterVersion, "0.0.0")
}

func TestGetVersion(t *testing.T) {
	actions := NewVersionController("foo-version")
	assert := assert.New(t)
	ctx := context.Background()
	response := actions.GetVersion(ctx)
	assert.NotEmpty(response)
	assert.Equal("foo-version", response.Etcd.Cluster)
	assert.Contains(response.Etcd.Server, "3")
	assert.Contains(response.SensuBackend, "#")
}
