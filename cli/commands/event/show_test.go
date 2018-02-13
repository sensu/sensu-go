package event

import (
	"fmt"
	"testing"

	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShowCommand(t *testing.T) {
	cli := test.NewMockCLI()
	cli.Config.(*client.MockConfig).On("Format").Return("json")
	cmd := ShowCommand(cli)

	assert.NotNil(t, cmd, "cmd should be returned")
	assert.NotNil(t, cmd.RunE, "cmd should be able to be executed")
	assert.Regexp(t, "info", cmd.Use)
	assert.Regexp(t, "event", cmd.Short)
}

func TestShowCommandRunEClosure(t *testing.T) {
	cli := test.NewMockCLI()
	cli.Client.(*client.MockClient).
		On("FetchEvent", "foo", "check_foo").
		Return(types.FixtureEvent("foo", "check_foo"), nil)
	cli.Config.(*client.MockConfig).On("Format").Return("json")

	cmd := ShowCommand(cli)
	out, err := test.RunCmd(cmd, []string{"foo", "check_foo"})

	assert.NotEmpty(t, out)
	assert.Contains(t, out, "check_foo")
	assert.Nil(t, err)
}

func TestShowCommandRunMissingArgs(t *testing.T) {
	cli := test.NewMockCLI()
	cli.Config.(*client.MockConfig).On("Format").Return("json")
	cmd := ShowCommand(cli)
	out, err := test.RunCmd(cmd, []string{})
	require.Error(t, err)
	assert.NotEmpty(t, out)
	assert.Contains(t, out, "Usage")
}

func TestShowCommandRunEClosureWithTable(t *testing.T) {
	cli := test.NewMockCLI()
	cli.Client.(*client.MockClient).
		On("FetchEvent", "foo", "check_foo").
		Return(types.FixtureEvent("foo", "check_foo"), nil)
	cli.Config.(*client.MockConfig).On("Format").Return("tabular")

	cmd := ShowCommand(cli)
	require.NoError(t, cmd.Flags().Set("format", "tabular"))

	out, err := test.RunCmd(cmd, []string{"foo", "check_foo"})
	require.NoError(t, err)
	assert.NotEmpty(t, out)
	assert.Contains(t, out, "Entity")
	assert.Contains(t, out, "Check")
}

func TestShowCommandRunEClosureWithErr(t *testing.T) {
	cli := test.NewMockCLI()
	cli.Client.(*client.MockClient).
		On("FetchEvent", "foo", "check_foo").
		Return(types.FixtureEvent("foo", "check_foo"), fmt.Errorf("error"))
	cli.Config.(*client.MockConfig).On("Format").Return("json")

	cmd := ShowCommand(cli)
	out, err := test.RunCmd(cmd, []string{"foo", "check_foo"})

	assert.NotNil(t, err)
	assert.Equal(t, "error", err.Error())
	assert.Empty(t, out)
}
