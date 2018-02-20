package environment

import (
	"fmt"
	"testing"

	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/sensu/sensu-go/types"
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
		{[]string{"foo"}, nil, fmt.Errorf("error"), "", true},
	}

	for _, tc := range testCases {
		name := ""
		if len(tc.args) > 0 {
			name = tc.args[0]
		}

		testName := fmt.Sprintf(
			"update the environment %s",
			name,
		)
		t.Run(testName, func(t *testing.T) {
			env := types.FixtureEnvironment(name)
			cli := test.NewMockCLI()

			client := cli.Client.(*client.MockClient)
			client.On(
				"FetchEnvironment",
				name,
			).Return(env, tc.fetchResponse)

			client.On(
				"UpdateEnvironment",
				"default",
				mock.Anything,
			).Return(tc.updateResponse)

			cmd := UpdateCommand(cli)
			out, err := test.RunCmd(cmd, tc.args)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Regexp(t, tc.expectedOutput, out)
		})
	}
}
