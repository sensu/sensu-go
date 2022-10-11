package handler

import (
	"fmt"
	"testing"

	"github.com/sensu/sensu-go/cli"
	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	corev2 "github.com/sensu/core/v2"
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
			"update the handler %s",
			name,
		)
		t.Run(testName, func(t *testing.T) {
			test.WithMockCLI(t, func(cli *cli.SensuCli) {
				handler := corev2.FixtureHandler(name)

				client := cli.Client.(*client.MockClient)
				client.On(
					"FetchHandler",
					name,
				).Return(handler, tc.fetchResponse)

				client.On(
					"CreateHandler",
					mock.Anything,
				).Return(tc.updateResponse)

				cmd := UpdateCommand(cli)
				out, err := test.RunCmdWithOutFile(cmd, tc.args, cli.OutFile)

				assert.Regexp(t, tc.expectedOutput, out)
				if tc.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		})
	}
}
