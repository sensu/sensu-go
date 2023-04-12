package subcommands

import (
	"errors"
	"fmt"
	"os"
	"testing"

	v2 "github.com/sensu/core/v2"
	client "github.com/sensu/sensu-go/cli/client/testing"
	stest "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSetProxyRequestsCommand(t *testing.T) {
	const proxyJSON = `{"entity_attributes":["entity.EntityClass == \"proxy\""], "splay":true, "splay_coverage":90}`
	const invalidProxyJSON = `{"splay":true, "splay_coverage":200}`
	tests := []struct {
		args		[]string
		useflag		bool
		stdin		string
		fetchResponse	error
		updateResponse	error
		expectedOutput	string
		expectError	bool
	}{
		{[]string{}, false, "", nil, nil, "Usage", true},
		{[]string{"foo"}, false, "", errors.New("error"), nil, "", true},
		{[]string{"bar"}, false, "", nil, errors.New("error"), "", true},
		{[]string{"check1"}, false, "", nil, nil, "", true},
		{[]string{"check1"}, false, proxyJSON, nil, nil, "Updated", false},
		{[]string{"check1"}, false, "invalidjson", nil, nil, "", true},
		{[]string{"check1"}, true, proxyJSON, nil, nil, "", false},
		{[]string{"check1"}, true, invalidProxyJSON, nil, nil, "", true},
	}

	for i, test := range tests {
		name := ""
		if len(test.args) > 0 {
			name = test.args[0]
		}
		t.Run(fmt.Sprintf("test %d", i), func(t *testing.T) {
			check := v2.FixtureCheckConfig("check1")
			cli := stest.NewMockCLI()
			client := cli.Client.(*client.MockClient)
			client.On("FetchCheck", name).Return(check, test.fetchResponse)
			client.On("UpdateCheck", mock.Anything).Return(test.updateResponse)
			cmd := SetProxyRequestsCommand(cli)
			name, stdin, cleanup := fileFromString(t, test.stdin)
			defer cleanup()
			if test.useflag {
				require.NoError(t, stdin.Close())
				require.NoError(t, cmd.Flags().Set("file", name))
			} else {
				os.Stdin = stdin
			}
			out, err := stest.RunCmd(cmd, test.args)
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Regexp(t, test.expectedOutput, out)
		})
	}
}
