package basic

import (
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestAPIUrl(t *testing.T) {
	config := &Config{Cluster: Cluster{APIUrl: "localhost"}}
	assert.Equal(t, config.Cluster.APIUrl, config.APIUrl())
}

func TestFormat(t *testing.T) {
	config := &Config{Profile: Profile{Format: "json"}}
	assert.Equal(t, config.Profile.Format, config.Format())
}

func TestTokens(t *testing.T) {
	tokens := &types.Tokens{Access: "foobar"}
	config := &Config{Cluster: Cluster{Tokens: tokens}}
	assert.Equal(t, tokens.Access, config.Tokens().Access)
}
