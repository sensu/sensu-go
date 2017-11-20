package event

import (
	"fmt"
	"testing"

	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/stretchr/testify/assert"
)

func TestDeleteCommand(t *testing.T) {
	cli := test.NewMockCLI()
	cmd := DeleteCommand(cli)

	assert.NotNil(t, cmd, "cmd should be returned")
	assert.NotNil(t, cmd.RunE, "cmd should be able to be executed")
	assert.Regexp(t, "delete", cmd.Use)
	assert.Regexp(t, "event", cmd.Short)
}

func TestDeleteCommandRunEClosure(t *testing.T) {
	cli := test.NewMockCLI()
	cli.Client.(*client.MockClient).
		On("DeleteEvent", "foo", "check_foo").
		Return(nil)

	cmd := DeleteCommand(cli)
	cmd.Flags().Set("skip-confirm", "t")
	out, err := test.RunCmd(cmd, []string{"foo", "check_foo"})

	assert.NotEmpty(t, out)
	assert.Contains(t, out, "Deleted")
	assert.Nil(t, err)
}

func TestDeleteCommandRunMissingArgs(t *testing.T) {
	cli := test.NewMockCLI()
	cmd := DeleteCommand(cli)
	cmd.Flags().Set("skip-confirm", "t")
	out, err := test.RunCmd(cmd, []string{})

	assert.NotEmpty(t, out)
	assert.Contains(t, out, "Usage")
	assert.NotNil(t, err)
}

func TestDeleteCommandRunEClosureWithErr(t *testing.T) {
	cli := test.NewMockCLI()
	cli.Client.(*client.MockClient).
		On("DeleteEvent", "foo", "check_foo").
		Return(fmt.Errorf("error"))

	cmd := DeleteCommand(cli)
	cmd.Flags().Set("skip-confirm", "t")
	out, err := test.RunCmd(cmd, []string{"foo", "check_foo"})

	assert.NotNil(t, err)
	assert.Equal(t, "error", err.Error())
	assert.Empty(t, out)
}

func TestDeleteCommandRunEFailConfirm(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := DeleteCommand(cli)
	out, err := test.RunCmd(cmd, []string{"foo", "check_foo"})

	assert.Contains(out, "Canceled")
	assert.NoError(err)
}
