package environment

import (
	"fmt"
	"testing"

	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateCommand(t *testing.T) {
	testCases := []struct {
		name           string
		storeResponse  error
		expectedOutput string
		expectError    bool
	}{
		{"", nil, "", true},
		{"foo", fmt.Errorf("error"), "", true},
		{"foo", nil, "Created", false},
	}

	for _, tc := range testCases {
		testName := fmt.Sprintf(
			"create the environment %s",
			tc.name,
		)
		t.Run(testName, func(t *testing.T) {
			cli := test.NewMockCLI()

			config := cli.Config.(*client.MockConfig)
			config.On("Organization").Return("default")

			client := cli.Client.(*client.MockClient)
			client.On(
				"CreateEnvironment",
				"default",
				mock.AnythingOfType("*types.Environment"),
			).Return(tc.storeResponse)

			cmd := CreateCommand(cli)
			cmd.Flags().Set("description", tc.name)
			out, err := test.RunCmd(cmd, []string{tc.name})

			assert.Regexp(t, tc.expectedOutput, out)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
