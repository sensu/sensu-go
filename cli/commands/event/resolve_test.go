package event

import (
	"fmt"
	"testing"

	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	corev2 "github.com/sensu/core/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestResolveCommand(t *testing.T) {
	testCases := []struct {
		args           []string
		fetchResponse  error
		updateResponse error
		expectedOutput string
		expectError    bool
	}{
		{[]string{}, nil, nil, "Usage", true},
		{[]string{"foo", "bar"}, nil, nil, "", false},
		{[]string{"foo", "bar"}, fmt.Errorf("error"), nil, "", true},
		{[]string{"foo", "bar"}, nil, fmt.Errorf("error"), "", true},
	}

	for _, tc := range testCases {
		name := ""
		if len(tc.args) > 0 {
			name = tc.args[0]
		}

		testName := fmt.Sprintf(
			"resolve the event %s",
			name,
		)
		t.Run(testName, func(t *testing.T) {
			event := corev2.FixtureEvent("entity", "check")
			cli := test.NewMockCLI()

			client := cli.Client.(*client.MockClient)
			client.On(
				"FetchEvent",
				"foo", "bar",
			).Return(event, tc.fetchResponse)

			client.On(
				"ResolveEvent",
				mock.Anything,
			).Return(tc.updateResponse)

			cmd := ResolveCommand(cli)
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
