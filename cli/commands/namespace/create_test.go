package namespace

import (
	"testing"

	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateCommand(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := CreateCommand(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("create", cmd.Use)
	assert.Regexp("namespace", cmd.Short)
}

func TestCreateCommandRunEClosure(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewMockCLI()
	cli.Client.(*client.MockClient).
		On("CreateNamespace", mock.Anything).
		Return(nil)

	cmd := CreateCommand(cli)
	out, err := test.RunCmd(cmd, []string{"foo"})

	assert.Regexp("Created", out)
	assert.NoError(err)
}
