package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseResourceKey(t *testing.T) {
	const input = "/sensu.io/checks/acme/baz"
	sk := ParseResourceKey(input)
	assert.Equal(t, "sensu.io", sk.Root)
	assert.Equal(t, "checks", sk.ResourceType)
	assert.Equal(t, "acme", sk.Namespace)
	assert.Equal(t, "baz", sk.ResourceName)
	assert.Equal(t, input, sk.String())
}

func TestParseResourceKeySmoke(t *testing.T) {
	ParseResourceKey("")
	ParseResourceKey("alsdkjflsakdjfl!!! /.cxlkj o2")
	ParseResourceKey("/")
	ParseResourceKey("///")
}
