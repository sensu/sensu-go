package configure

import (
	"testing"

	"github.com/sensu/sensu-go/cli/client/config"
	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"

	"github.com/stretchr/testify/assert"
)

func TestCommand(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	mockConfig := cli.Config.(*client.MockConfig)
	mockConfig.On("Format").Return(config.DefaultFormat)
	mockConfig.On("APIUrl").Return("http://127.0.0.1:8080")

	cmd := Command(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("configure", cmd.Use)
	assert.Regexp("Initialize sensuctl configuration", cmd.Short)
}
