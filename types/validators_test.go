package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateHandlerType(t *testing.T) {
	assert.Error(t, validateHandlerType(""))
	assert.Error(t, validateHandlerType("foo"))
	assert.NoError(t, validateHandlerType("pipe"))
	assert.NoError(t, validateHandlerType("tcp"))
	assert.NoError(t, validateHandlerType("udp"))
	assert.NoError(t, validateHandlerType("transport"))
	assert.NoError(t, validateHandlerType("set"))
}

func TestValidateName(t *testing.T) {
	assert.Error(t, validateName(""))
	assert.Error(t, validateName("foo bar"))
	assert.Error(t, validateName("foo@bar"))
	assert.NoError(t, validateName("foo-bar"))
}
