package logging

import (
	"fmt"
	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/sensu/sensu-go/util/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetModuleLogLevel(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := GetModuleLogLevel(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("get", cmd.Use)
	assert.Regexp("log level", cmd.Short)
}

func TestGetModuleLogLevelCommandRunEClosureWithoutFlags(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)

	client.On("Get", mock.Anything, mock.Anything).Return(fmt.Errorf("error"))

	cmd := GetModuleLogLevel(cli)
	out, err := test.RunCmd(cmd, []string{"foo"})

	assert.NotNil(err)
	assert.Equal("error", err.Error())
	assert.Empty(out)
}

func TestGetModuleLogLevelCommandRunEmptyArgs(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)

	client.On("Get", mock.Anything, mock.Anything).Return(fmt.Errorf("error"))

	cmd := GetModuleLogLevel(cli)
	out, err := test.RunCmd(cmd, []string{})

	assert.NotNil(err)
	assert.Regexp("argument", err.Error())
	assert.NotEmpty(out)
}

func TestGetModuleLogLevelRunEClosureWithTable(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewCLI()
	client := cli.Client.(*client.MockClient)
	var result logging.LogLevelRequest
	client.On("Get", mock.Anything, &result).Return(nil)

	cmd := GetModuleLogLevel(cli)
	require.NoError(t, cmd.Flags().Set("format", "tabular"))

	out, err := test.RunCmd(cmd, []string{"apid"})
	require.NoError(t, err)

	assert.NotEmpty(out)
	assert.Contains(out, "Module")
	assert.Contains(out, "Log Level")
}
