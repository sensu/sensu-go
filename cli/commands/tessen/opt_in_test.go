package tessen

import (
	"fmt"
	"testing"

	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOptInCommand(t *testing.T) {
	testCases := []struct {
		testName       string
		args           []string
		updateResponse error
		expectedOutput string
		expectError    bool
	}{
		{"args", []string{"foo"}, nil, "Usage", true},
		{"update error", []string{}, fmt.Errorf("error"), "", true},
		{"valid input", []string{}, nil, "Thank you so much for opting in!", false},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			cli := test.NewMockCLI()

			client := cli.Client.(*client.MockClient)
			client.On("Put", mock.Anything, mock.Anything).Return(tc.updateResponse)

			cmd := OptInCommand(cli)
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
