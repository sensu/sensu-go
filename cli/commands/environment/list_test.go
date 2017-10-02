package environment

import (
	"fmt"
	"testing"

	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestListCommand(t *testing.T) {
	type storeResponse struct {
		envs []types.Environment
		err  error
	}
	testCases := []struct {
		storeResponse  storeResponse
		format         string
		expectedOutput string
		expectError    bool
	}{
		{storeResponse{[]types.Environment{}, fmt.Errorf("error")}, "", "", true},
		{storeResponse{
			[]types.Environment{*types.FixtureEnvironment("one"), *types.FixtureEnvironment("two")},
			nil,
		}, "none", "Description", false},
		{storeResponse{
			[]types.Environment{*types.FixtureEnvironment("one"), *types.FixtureEnvironment("two")},
			nil,
		}, "json", "description", false},
	}

	for i, tc := range testCases {
		testName := fmt.Sprintf("list environments, test case #%d", i+1)
		t.Run(testName, func(t *testing.T) {
			cli := test.NewMockCLI()
			cli.Config.(*client.MockConfig).On("Format").Return(tc.format)

			client := cli.Client.(*client.MockClient)
			client.On(
				"ListEnvironments",
				"default",
			).Return(tc.storeResponse.envs, tc.storeResponse.err)

			cmd := ListCommand(cli)
			out, err := test.RunCmd(cmd, []string{})

			assert.Regexp(t, tc.expectedOutput, out)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
