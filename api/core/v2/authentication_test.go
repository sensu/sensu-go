package v2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixtureCreatesValidTokens(t *testing.T) {
	tokens := FixtureTokens("foo", "bar")
	assert.NoError(t, tokens.Validate())
}

func TestTokensValidator(t *testing.T) {
	tokens := FixtureTokens("foo", "bar")

	// Given valid tokens it should pass
	assert.NoError(t, tokens.Validate())

	// Given asset without an access token it should not pass
	tokens = FixtureTokens("", "foo")
	assert.Error(t, tokens.Validate())

	// Given asset without a refresh token it should not pass
	tokens = FixtureTokens("foo", "")
	assert.Error(t, tokens.Validate())

	// Given asset without an expiration it should not pass
	tokens = FixtureTokens("foo", "")
	tokens.ExpiresAt = 0
	assert.Error(t, tokens.Validate())
}
