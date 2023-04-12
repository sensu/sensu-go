package silenced

import (
	"errors"
	"strings"
	"testing"
	"time"

	v2 "github.com/sensu/core/v2"
	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
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
	client.On("FetchSilenced", mock.Anything).Return(v2.FixtureSilenced("foo:bar"), nil)

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
	client.On("FetchSilenced", mock.Anything).Return(v2.FixtureSilenced("foo:bar"), nil)

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
	client.On("FetchSilenced", mock.Anything).Return(&v2.Silenced{}, errors.New("my-err"))

	cmd := InfoCommand(cli)
	out, err := test.RunCmd(cmd, []string{"foo:bar"})

	assert.Equal("my-err", err.Error())
	assert.Empty(out)
}

func Test_expireAt(t *testing.T) {
	unixToInternal := int64((1969*365 + 1969/4 - 1969/100 + 1969/400) * 24 * 60 * 60)
	future := time.Unix(1<<63-1-unixToInternal, 999999999)

	tests := []struct {
		name		string
		expireAt	int64
		want		string
	}{
		{
			name:		"an entry without an expiration returns -1",
			expireAt:	-1,
			want:		"no expiration",
		},
		{
			name:		"an RFC3339 date is returned",
			expireAt:	future.Unix(),
			want:		"292277024627-12",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expireAt(tt.expireAt)
			if !strings.Contains(got, tt.want) {
				t.Errorf("expireAt() = %v, want %v", got, tt.want)
			}
		})
	}
}
