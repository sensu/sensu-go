package environment

import (
	"fmt"
	"testing"

	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteCommand(t *testing.T) {
	testCases := []struct {
		name           string
		storeResponse  error
		expectedOutput string
		expectError    bool
		skipConfirm    bool
	}{
		{"", nil, "Usage", true, true},
		{"foo", fmt.Errorf("error"), "", true, true},
		{"foo", nil, "Deleted", false, true},
		{"foo", nil, "Canceled", false, false},
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
			if tc.skipConfirm {
				require.NoError(t, cmd.Flags().Set("skip-confirm", "t"))
			}

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
