package asset

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
)

// Match a tabular output header
var tabularHeaderPattern = ".*\n( |─)+\n"

func TestOutdatedCommand(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewMockCLI()

	config := cli.Config.(*client.MockConfig)
	config.On("Format").Return("none")

	assets := []corev2.Asset{}
	client := cli.Client.(*client.MockClient)
	client.On("List", mock.Anything, &assets, mock.Anything, mock.Anything).Return(nil)

	cmd := OutdatedCommand(cli)
	out, err := test.RunCmd(cmd, []string{})

	// Match a tabular output header with nothing else under it
	assert.Regexp(tabularHeaderPattern+"$", out)
	assert.Nil(err)
}
