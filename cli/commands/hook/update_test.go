package hook

import (
	"fmt"
	"testing"

	v2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/cli"
	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateCommand(t *testing.T) {
	testCases := []struct {
		args           []string
		fetchResponse  error
		updateResponse error
		expectedOutput string
		expectError    bool
	}{
		{[]string{}, nil, nil, "Usage", true},
		{[]string{"foo"}, fmt.Errorf("error"), nil, "", true},
		{[]string{"bar"}, nil, fmt.Errorf("error"), "", true},
	}

	for _, tc := range testCases {
		name := ""
		if len(tc.args) > 0 {
			name = tc.args[0]
		}

		testName := fmt.Sprintf(
			"update the hook %s",
			name,
		)
		t.Run(testName, func(t *testing.T) {
			test.WithMockCLI(t, func(cli *cli.SensuCli) {
				hook := v2.FixtureHookConfig("my-id")

				client := cli.Client.(*client.MockClient)
				client.On(
					"FetchHook",
					name,
				).Return(hook, tc.fetchResponse)

				client.On(
					"UpdateHook",
					mock.Anything,
				).Return(tc.updateResponse)

				cmd := UpdateCommand(cli)
				out, err := test.RunCmdWithOutFile(cmd, tc.args, cli.OutFile)
				if tc.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}

				assert.Regexp(t, tc.expectedOutput, out)
			})
		})
	}
}
