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

//
// import (
// 	"errors"
// 	"testing"
//
// 	client "github.com/sensu/sensu-go/cli/client/testing"
// 	test "github.com/sensu/sensu-go/cli/commands/testing"
// 	"github.com/stretchr/testify/assert"
// )
//
// func TestDeleteCommand(t *testing.T) {
// 	assert := assert.New(t)
//
// 	cli := test.NewMockCLI()
// 	cmd := DeleteCommand(cli)
//
// 	assert.NotNil(cmd, "cmd should be returned")
// 	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
// 	assert.Regexp("delete", cmd.Use)
// 	assert.Regexp("organization", cmd.Short)
// }
//
// func TestDeleteCommandRunEClosureWithoutName(t *testing.T) {
// 	assert := assert.New(t)
//
// 	cli := test.NewMockCLI()
// 	cmd := DeleteCommand(cli)
// 	cmd.Flags().Set("timeout", "15")
// 	out, err := test.RunCmd(cmd, []string{})
//
// 	assert.Regexp("Usage", out) // usage should print out
// 	assert.Nil(err)
// }
//
// func TestDeleteCommandRunEClosureWithFlags(t *testing.T) {
// 	assert := assert.New(t)
//
// 	cli := test.NewMockCLI()
// 	client := cli.Client.(*client.MockClient)
// 	client.On("DeleteOrganization", "foo").Return(nil)
//
// 	cmd := DeleteCommand(cli)
// 	out, err := test.RunCmd(cmd, []string{"foo"})
//
// 	assert.Regexp("Deleted", out)
// 	assert.Nil(err)
// }
//
// func TestDeleteCommandRunEClosureWithServerErr(t *testing.T) {
// 	assert := assert.New(t)
//
// 	cli := test.NewMockCLI()
// 	client := cli.Client.(*client.MockClient)
// 	client.On("DeleteOrganization", "bar").Return(errors.New("oh noes"))
//
// 	cmd := DeleteCommand(cli)
// 	out, err := test.RunCmd(cmd, []string{"bar"})
//
// 	assert.Empty(out)
// 	assert.NotNil(err)
// 	assert.Equal("oh noes", err.Error())
// }
