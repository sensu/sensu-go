package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateHandlerType(t *testing.T) {
	assert.NotNil(t, validateHandlerType(""))
	assert.NotNil(t, validateHandlerType("foo"))
	assert.Nil(t, validateHandlerType("pipe"))
	assert.Nil(t, validateHandlerType("tcp"))
	assert.Nil(t, validateHandlerType("udp"))
	assert.Nil(t, validateHandlerType("transport"))
	assert.Nil(t, validateHandlerType("set"))
}

func TestValidateName(t *testing.T) {
	assert.NotNil(t, validateName(""))
	assert.NotNil(t, validateName("foo bar"))
	assert.NotNil(t, validateName("foo@bar"))
	assert.Nil(t, validateName("foo-bar"))
}
