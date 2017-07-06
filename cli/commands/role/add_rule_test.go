package role

import (
	"testing"

	clientmock "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAddRuleCommand(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := AddRuleCommand(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("add-rule", cmd.Use)
	assert.Regexp("to role", cmd.Short)
}

func TestAddRuleCommandRunEClosureSucess(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewMockCLI()

	client := cli.Client.(*clientmock.MockClient)
	client.On("AddRule", "name", mock.AnythingOfType("*types.Rule")).Return(nil)

	cmd := AddRuleCommand(cli)
	cmd.Flags().Set("type", "*")
	cmd.Flags().Set("create", "t")

	out, err := test.RunCmd(cmd, []string{"name"})

	assert.Contains(out, "Added")
	assert.NoError(err)
}

func TestAddRuleCommandRunEInvalid(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewMockCLI()

	cmd := AddRuleCommand(cli)
	out, err := test.RunCmd(cmd, []string{"name"})

	assert.Empty(out)
	assert.Error(err)
}

func TestAddRuleCommandRunEClosureServerErr(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewMockCLI()

	client := cli.Client.(*clientmock.MockClient)
	client.On("AddRule", "name", mock.AnythingOfType("*types.rule")).Return(nil)

	cmd := AddRuleCommand(cli)
	out, err := test.RunCmd(cmd, []string{"name"})

	assert.Empty(out)
	assert.Error(err)
}

func TestAddRuleCommandRunEClosureMissingArgs(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewMockCLI()

	cmd := AddRuleCommand(cli)
	out, err := test.RunCmd(cmd, []string{})

	assert.Empty(out)
	assert.Error(err)
}
