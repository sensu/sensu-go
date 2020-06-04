package user

import (
	"testing"

	v2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSetPasswordCommand(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := SetPasswordCommand(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("change", cmd.Use)
	assert.Regexp("change password", cmd.Short)
}

func TestSetPasswordCommandRunEClosure(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	clientMock := cli.Client.(*client.MockClient)
	configMock := cli.Config.(*client.MockConfig)
	clientMock.On("UpdatePassword", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	claims := v2.FixtureClaims("foo", nil)
	_, tokenString, _ := jwt.AccessToken(claims)
	configMock.On("Tokens").Return(&v2.Tokens{Access: tokenString})

	cmd := SetPasswordCommand(cli)
	require.NoError(t, cmd.Flags().Set("interactive", "false"))
	require.NoError(t, cmd.Flags().Set("current-password", "my-new-password"))
	require.NoError(t, cmd.Flags().Set("new-password", "my-new-password"))
	out, err := test.RunCmd(cmd, []string{"my-username"})
	assert.NoError(err)
	assert.Regexp("Updated", out)
}
