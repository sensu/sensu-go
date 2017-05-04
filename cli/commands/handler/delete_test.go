package handler

import (
	"errors"
	"testing"

	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDeleteCommand(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := DeleteCommand(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("delete", cmd.Use)
	assert.Regexp("handlers", cmd.Short)
}

func TestDeleteCommandRunEClosureWithoutName(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := DeleteCommand(cli)
	cmd.Flags().Set("timeout", "15")
	out, err := test.RunCmd(cmd, []string{})

	assert.Regexp("Usage", out) // usage should print out
	assert.Nil(err)
}

func TestDeleteCommandRunEClosureWithFlags(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("DeleteHandler", mock.AnythingOfType("*types.Handler")).Return(nil)

	cmd := DeleteCommand(cli)
	out, err := test.RunCmd(cmd, []string{"test-handler"})

	assert.Regexp("OK", out)
	assert.Nil(err)
}

func TestDeleteCommandRunEClosureWithServerErr(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("DeleteHandler", mock.AnythingOfType("*types.Handler")).Return(errors.New("oh noes"))

	cmd := DeleteCommand(cli)
	out, err := test.RunCmd(cmd, []string{"test-handler"})

	assert.Empty(out)
	assert.NotNil(err)
	assert.Equal("oh noes", err.Error())
}
