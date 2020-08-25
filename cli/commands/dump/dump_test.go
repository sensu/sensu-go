package dump

import (
	"errors"
	"testing"

	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/stretchr/testify/assert"
)

func TestCommand(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewCLI()
	cmd := Command(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("dump", cmd.Use)
}

func TestCommandArgs(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewCLI()
	cmd := Command(cli)

	out, err := test.RunCmd(cmd, []string{})
	assert.NotEmpty(out)
	assert.Error(err)

	// invalid resources
	out, err = test.RunCmd(cmd, []string{"check,foo"})
	assert.Empty(out)
	assert.Error(err)
}

func TestListFlags(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewCLI()
	cmd := Command(cli)

	flag := cmd.Flag("all-namespaces")
	assert.NotNil(flag)

	flag = cmd.Flag("format")
	assert.NotNil(flag)

	flag = cmd.Flag("file")
	assert.NotNil(flag)
}

func TestGetAcceptedResourceTypes(t *testing.T) {
	assert := assert.New(t)
	accepted := getAcceptedResourceTypes()
	assert.NotEmpty(accepted)
	assert.Contains(accepted, "all")
	assert.Contains(accepted, "checks")
	assert.Contains(accepted, "core/v2.CheckConfig")
	assert.NotContains(accepted, "foo")
}

func TestCheckArgs(t *testing.T) {
	assert := assert.New(t)
	accepted := getAcceptedResourceTypes()
	tests := []struct {
		name string
		args string
		err  error
	}{
		{
			name: "valid all keyword",
			args: "all",
			err:  nil,
		},
		{
			name: "valid short name",
			args: "checks",
			err:  nil,
		},
		{
			name: "valid full name",
			args: "core/v2.CheckConfig",
			err:  nil,
		},
		{
			name: "empty args",
			args: "",
			err:  nil,
		},
		{
			name: "invalid name",
			args: "foo",
			err:  errors.New("invalid resource type: foo"),
		},
		{
			name: "valid multiple args",
			args: "all,checks,core/v2.CheckConfig",
			err:  nil,
		},
		{
			name: "invalid multiple args",
			args: "checks,foo",
			err:  errors.New("invalid resource type: foo"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(tc.err, checkArgs(tc.args, accepted))
		})
	}
}
