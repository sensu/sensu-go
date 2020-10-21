package silenced

import (
	"errors"
	"strings"
	"testing"
	"time"

	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestInfoCommand(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewCLI()
	cmd := InfoCommand(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("info", cmd.Use)
	assert.Regexp("silenced", cmd.Short)
}

func TestInfoCommandRunEClosure(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewCLI()
	client := cli.Client.(*client.MockClient)
	client.On("FetchSilenced", mock.Anything).Return(types.FixtureSilenced("foo:bar"), nil)

	cmd := InfoCommand(cli)
	out, err := test.RunCmd(cmd, []string{"foo:bar"})
	require.NoError(t, err)

	assert.NotEmpty(out)
	assert.Contains(out, "foo:bar")
}

func TestInfoCommandRunMissingArgs(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewCLI()
	cmd := InfoCommand(cli)
	out, err := test.RunCmd(cmd, []string{"wrong", "stuff"})

	require.Error(t, err)
	assert.NotEmpty(out)
	assert.Contains(out, "Usage")
}

func TestInfoCommandRunEClosureWithTable(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewCLI()
	client := cli.Client.(*client.MockClient)
	client.On("FetchSilenced", mock.Anything).Return(types.FixtureSilenced("foo:bar"), nil)

	cmd := InfoCommand(cli)
	require.NoError(t, cmd.Flags().Set("format", "tabular"))

	out, err := test.RunCmd(cmd, []string{"foo:bar"})
	require.NoError(t, err)

	assert.NotEmpty(out)
	assert.Contains(out, "Reason")
	assert.Contains(out, "Subscription")
	assert.Contains(out, "Namespace")
}

func TestInfoCommandRunEClosureWithErr(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewCLI()
	client := cli.Client.(*client.MockClient)
	client.On("FetchSilenced", mock.Anything).Return(&types.Silenced{}, errors.New("my-err"))

	cmd := InfoCommand(cli)
	out, err := test.RunCmd(cmd, []string{"foo:bar"})

	assert.Equal("my-err", err.Error())
	assert.Empty(out)
}

func Test_expireTime(t *testing.T) {
	tests := []struct {
		name          string
		beginTS       int64
		expireSeconds int64
		want          string
	}{
		{
			name:          "an entry without an expiration returns -1",
			beginTS:       time.Now().Truncate(time.Duration(1) * time.Minute).Unix(),
			expireSeconds: -1,
			want:          "-1",
		},
		{
			name:          "an entry that is not yet in effect return a RFC3339 date",
			beginTS:       time.Now().Add(time.Duration(1) * time.Minute).Unix(),
			expireSeconds: 300,
			want:          "2020-10-",
		},
		{
			name:          "an entry that is in effect return the configured duration",
			beginTS:       time.Now().Truncate(time.Duration(1) * time.Minute).Unix(),
			expireSeconds: 300,
			want:          "5m",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expireTime(tt.beginTS, tt.expireSeconds)
			if !strings.Contains(got, tt.want) {
				t.Errorf("expireTime() = %v, want %v", got, tt.want)
			}
		})
	}
}
