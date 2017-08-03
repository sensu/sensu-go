package environment

import (
	"fmt"
	"testing"

	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/stretchr/testify/assert"
)

func TestDeleteCommand(t *testing.T) {
	testCases := []struct {
		name           string
		storeResponse  error
		expectedOutput string
		expectError    bool
	}{
		{"", nil, "Usage", false},
		{"foo", fmt.Errorf("error"), "", true},
		{"foo", nil, "Deleted", false},
	}

	for _, tc := range testCases {
		testName := fmt.Sprintf(
			"delete the environment %s",
			tc.name,
		)
		t.Run(testName, func(t *testing.T) {
			cli := test.NewMockCLI()

			client := cli.Client.(*client.MockClient)
			client.On(
				"DeleteEnvironment",
				"default",
				tc.name,
			).Return(tc.storeResponse)

			cmd := DeleteCommand(cli)
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
