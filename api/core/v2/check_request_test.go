package v2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProxyRequestsValidate(t *testing.T) {
	var p ProxyRequests

	// Invalid splay coverage
	p.SplayCoverage = 150
	assert.Error(t, p.Validate())
	p.SplayCoverage = 0

	// Invalid splay and splay coverage
	p.Splay = true
	assert.Error(t, p.Validate())
	p.SplayCoverage = 90

	p.EntityAttributes = []string{`entity.EntityClass == "proxy"`}

	// Valid proxy request
	assert.NoError(t, p.Validate())
}

func TestFixtureProxyRequests(t *testing.T) {
	p := FixtureProxyRequests(true)

	assert.Equal(t, true, p.Splay)
	assert.Equal(t, uint32(90), p.SplayCoverage)
	assert.NoError(t, p.Validate())

	p.SplayCoverage = 0
	assert.Error(t, p.Validate())
}
