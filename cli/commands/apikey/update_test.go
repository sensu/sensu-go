package apikey

import (
	"fmt"
	"testing"

	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateCommand(t *testing.T) {
	testCases := []struct {
		args           []string
		updateResponse error
		expectedOutput string
		expectError    bool
	}{
		{[]string{}, nil, "Usage", true},
		{[]string{"my-api-key"}, fmt.Errorf("err"), "", true},
	}

	for _, tc := range testCases {
		name := ""
		if len(tc.args) > 0 {
			name = tc.args[0]
		}

		testName := fmt.Sprintf(
			"update the apikey %s",
			name,
		)
		t.Run(testName, func(t *testing.T) {
			cli := test.NewMockCLI()
			client := cli.Client.(*client.MockClient)
			client.On("Patch", mock.Anything, mock.Anything).Return(tc.updateResponse)

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
