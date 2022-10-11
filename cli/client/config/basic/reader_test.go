package basic

import (
	"testing"
	"time"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/cli/client/config"
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

func TestTimeout(t *testing.T) {
	conf := &Config{Cluster: Cluster{Timeout: 30 * time.Second}}
	assert.Equal(t, conf.Cluster.Timeout, conf.Timeout())
}

func TestTimeoutDefault(t *testing.T) {
	conf := &Config{}
	assert.Equal(t, config.DefaultTimeout, conf.Timeout())
}

func TestTokens(t *testing.T) {
	tokens := &corev2.Tokens{Access: "foobar"}
	conf := &Config{Cluster: Cluster{Tokens: tokens}}
	assert.Equal(t, tokens.Access, conf.Tokens().Access)
}

func TestAPIKey(t *testing.T) {
	apiKey := "f70f5a0a-c443-453e-8354-76e40b8b73bd"
	conf := &Config{Cluster: Cluster{APIKey: apiKey}}
	assert.Equal(t, apiKey, conf.APIKey())
}
