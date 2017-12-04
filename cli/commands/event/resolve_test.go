package event

import (
	"fmt"
	"testing"

	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/stretchr/testify/assert"
)

func TestResolveCommand(t *testing.T) {
	cli := test.NewMockCLI()
	cmd := ResolveCommand(cli)

	assert.NotNil(t, cmd, "cmd should be returned")
	assert.NotNil(t, cmd.RunE, "cmd should be able to be executed")
	assert.Regexp(t, "resolve", cmd.Use)
	assert.Regexp(t, "event", cmd.Short)
}

func TestResolveCommandRunEClosure(t *testing.T) {
	cli := test.NewMockCLI()
	cli.Client.(*client.MockClient).
		On("ResolveEvent", "foo", "check_foo").
		Return(nil)

	cmd := ResolveCommand(cli)
	out, err := test.RunCmd(cmd, []string{"foo", "check_foo"})

	assert.NotEmpty(t, out)
	assert.Contains(t, out, "OK")
	assert.Nil(t, err)
}

func TestResolveCommandRunMissingArgs(t *testing.T) {
	cli := test.NewMockCLI()
	cmd := ResolveCommand(cli)
	out, err := test.RunCmd(cmd, []string{})

	assert.NotEmpty(t, out)
	assert.Contains(t, out, "Usage")
	assert.NotNil(t, err)
}

func TestResolveCommandRunEClosureWithErr(t *testing.T) {
	cli := test.NewMockCLI()
	cli.Client.(*client.MockClient).
		On("ResolveEvent", "foo", "check_foo").
		Return(fmt.Errorf("error"))

	cmd := DeleteCommand(cli)
	cmd.Flags().Set("skip-confirm", "t")
	out, err := test.RunCmd(cmd, []string{"foo", "check_foo"})

	assert.NotNil(t, err)
	assert.Equal(t, "error", err.Error())
	assert.Empty(t, out)
}

func TestResolveCommandRunEFailConfirm(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := ResolveCommand(cli)
	out, err := test.RunCmd(cmd, []string{"foo", "check_foo"})

	assert.Contains(out, "Canceled")
	assert.NoError(err)
}
