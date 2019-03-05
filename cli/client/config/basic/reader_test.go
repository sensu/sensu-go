package basic

import (
	"testing"

	"github.com/sensu/sensu-go/cli/client/config"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestAPIUrl(t *testing.T) {
	conf := &Config{Cluster: Cluster{APIUrl: "localhost"}}
	assert.Equal(t, conf.Cluster.APIUrl, conf.APIUrl())
}

func TestFormat(t *testing.T) {
	conf := &Config{Profile: Profile{Format: "json"}}
	assert.Equal(t, conf.Profile.Format, conf.Format())
}

func TestFormatDefault(t *testing.T) {
	conf := &Config{}
	assert.Equal(t, config.DefaultFormat, conf.Format())
}

func TestNamespace(t *testing.T) {
	conf := &Config{Profile: Profile{Namespace: "dev"}}
	assert.Equal(t, conf.Profile.Namespace, conf.Namespace())
}

func TestNamespaceDefault(t *testing.T) {
	conf := &Config{}
	assert.Equal(t, config.DefaultNamespace, conf.Namespace())
}

func TestTokens(t *testing.T) {
	tokens := &types.Tokens{Access: "foobar"}
	conf := &Config{Cluster: Cluster{Tokens: tokens}}
	assert.Equal(t, tokens.Access, conf.Tokens().Access)
}
