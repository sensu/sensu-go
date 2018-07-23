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

func TestEdition(t *testing.T) {
	conf := &Config{Cluster: Cluster{Edition: types.CoreEdition}}
	assert.Equal(t, types.CoreEdition, conf.Edition())

	conf.Cluster.Edition = ""
	assert.Equal(t, types.CoreEdition, conf.Edition())
}

func TestEnvironment(t *testing.T) {
	conf := &Config{Profile: Profile{Environment: "dev"}}
	assert.Equal(t, conf.Profile.Environment, conf.Environment())
}

func TestEnvironmentDefault(t *testing.T) {
	conf := &Config{}
	assert.Equal(t, config.DefaultEnvironment, conf.Environment())
}

func TestFormat(t *testing.T) {
	conf := &Config{Profile: Profile{Format: "json"}}
	assert.Equal(t, conf.Profile.Format, conf.Format())
}

func TestFormatDefault(t *testing.T) {
	conf := &Config{}
	assert.Equal(t, config.DefaultFormat, conf.Format())
}

func TestOrganization(t *testing.T) {
	conf := &Config{Profile: Profile{Organization: "dev"}}
	assert.Equal(t, conf.Profile.Organization, conf.Organization())
}

func TestOrganizationDefault(t *testing.T) {
	conf := &Config{}
	assert.Equal(t, config.DefaultOrganization, conf.Organization())
}

func TestTokens(t *testing.T) {
	tokens := &types.Tokens{Access: "foobar"}
	conf := &Config{Cluster: Cluster{Tokens: tokens}}
	assert.Equal(t, tokens.Access, conf.Tokens().Access)
}
