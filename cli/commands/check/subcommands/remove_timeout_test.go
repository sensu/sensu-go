package subcommands

import (
	"errors"
	"fmt"
	"testing"

	client "github.com/sensu/sensu-go/cli/client/testing"
	stest "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRemoveTimeoutCommand(t *testing.T) {
	tests := []struct {
		args           []string
		fetchResponse  error
		updateResponse error
		expectedOutput string
		expectError    bool
	}{
		{[]string{}, nil, nil, "Usage", true},
		{[]string{"foo"}, errors.New("error"), nil, "", true},
		{[]string{"bar"}, nil, errors.New("error"), "", true},
		{[]string{"check1"}, nil, nil, "OK", false},
	}

	for i, test := range tests {
		name := ""
		if len(test.args) > 0 {
			name = test.args[0]
		}
		t.Run(fmt.Sprintf("test %d", i), func(t *testing.T) {
			check := types.FixtureCheckConfig("check1")
			cli := stest.NewMockCLI()
			client := cli.Client.(*client.MockClient)
			client.On("FetchCheck", name).Return(check, test.fetchResponse)
			client.On("UpdateCheck", mock.Anything).Return(test.updateResponse)
			cmd := RemoveTimeoutCommand(cli)
			out, err := stest.RunCmd(cmd, test.args)
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Regexp(t, test.expectedOutput, out)
		})
	}
}
