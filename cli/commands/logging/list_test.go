package logging

import (
	"encoding/json"
	"errors"
	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/sensu/sensu-go/util/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestListLogLevelsCommand(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewCLI()
	cmd := ListLogLevels(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("list", cmd.Use)
	assert.Regexp("modules", cmd.Short)
}

func TestListLogLevelsCommandRunEClosure(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewCLI()
	client := cli.Client.(*client.MockClient)
	logEntry := logging.LogLevelRequest{
		Module: "Name",
		Level:  "info",
	}
	client.On("List", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(
		func(args mock.Arguments) {
			resources := args[1].(*[]logging.LogLevelRequest)
			*resources = []logging.LogLevelRequest{
				logEntry,
			}
		},
	)

	cmd := ListLogLevels(cli)
	out, err := test.RunCmd(cmd, []string{})

	assert.NotEmpty(out)
	assert.Nil(err)
	assert.NotContains(out, "==")
	assert.Contains(out, "Name")
}

func TestListLogLevelsCommandRunEClosureWithErr(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewCLI()
	client := cli.Client.(*client.MockClient)
	//resources := []corev2.User{}
	client.On("List", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("dunno"))

	cmd := ListLogLevels(cli)
	out, err := test.RunCmd(cmd, []string{})

	assert.Empty(out)
	assert.Error(err)
}

func TestListLogLevelsCommandRunEClosureWithTable(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewCLI()

	logEntry := logging.LogLevelRequest{
		Module: "Name",
		Level:  "info",
	}

	client := cli.Client.(*client.MockClient)
	client.On("List", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(
		func(args mock.Arguments) {
			resources := args[1].(*[]logging.LogLevelRequest)
			*resources = []logging.LogLevelRequest{
				logEntry,
			}
		},
	)

	cmd := ListLogLevels(cli)
	require.NoError(t, cmd.Flags().Set("format", "none"))
	out, err := test.RunCmd(cmd, []string{})

	assert.NotEmpty(out)
	assert.Contains(out, "Module")
	assert.Contains(out, "Level")
	assert.NoError(err)
}

func TestListLogLevelsCommandRunEClosureWithJSONOutput(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewCLI()

	logEntry := logging.LogLevelRequest{
		Module: "Name",
		Level:  "info",
	}
	logEntries := []logging.LogLevelRequest{
		logEntry,
	}

	expected, err := json.Marshal(logEntries)
	if err != nil {
		t.Fatal(err)
	}

	client := cli.Client.(*client.MockClient)
	client.On("List", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(
		func(args mock.Arguments) {
			resources := args[1].(*[]logging.LogLevelRequest)
			*resources = logEntries
		},
	)

	cmd := ListLogLevels(cli)
	require.NoError(t, cmd.Flags().Set("format", "json"))
	out, err := test.RunCmd(cmd, []string{})

	assert.NoError(err)
	assert.NotEmpty(out)
	assert.JSONEq(string(expected), out)
}
