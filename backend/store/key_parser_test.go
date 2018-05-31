package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseResourceKey(t *testing.T) {
	const input = "/sensu.io/checks/default/foobar/baz"
	sk := ParseResourceKey(input)
	assert.Equal(t, "sensu.io", sk.Root)
	assert.Equal(t, "checks", sk.ResourceType)
	assert.Equal(t, "default", sk.Organization)
	assert.Equal(t, "foobar", sk.Environment)
	assert.Equal(t, "baz", sk.ResourceName)
	assert.Equal(t, input, sk.String())
}

func TestParseResourceKeySmoke(t *testing.T) {
	ParseResourceKey("")
	ParseResourceKey("alsdkjflsakdjfl!!! /.cxlkj o2")
	ParseResourceKey("/")
	ParseResourceKey("///")
}
