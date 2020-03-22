package configure

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/sensu/sensu-go/cli/client/config"
	client "github.com/sensu/sensu-go/cli/client/testing"
	"github.com/sensu/sensu-go/cli/commands/root"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/sensu/sensu-go/types"

	"github.com/stretchr/testify/assert"
)

func TestCommand(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	mockConfig := cli.Config.(*client.MockConfig)
	mockConfig.On("Format").Return(config.DefaultFormat)
	mockConfig.On("APIUrl").Return("http://127.0.0.1:8080")
	mockConfig.On("Timeout").Return(time.Second * 15)

	cmd := Command(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("configure", cmd.Use)
	assert.Regexp("Initialize sensuctl configuration", cmd.Short)
}

func TestCommandRunEClosureWithFlags(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	mockClient := cli.Client.(*client.MockClient)
	mockConfig := cli.Config.(*client.MockConfig)
	mockConfig.On("APIUrl").Return("http://127.0.0.1:8080")
	mockConfig.On("Format").Return(config.DefaultFormat)
	mockConfig.On("SaveAPIUrl", mock.Anything).Return(nil)
	mockClient.On("CreateAccessToken", mock.Anything, mock.Anything, mock.Anything).Return(&types.Tokens{}, nil)
	mockConfig.On("SaveTokens", mock.Anything).Return(nil)
	mockConfig.On("SaveFormat", mock.Anything).Return(nil)
	mockClient.On("FetchUser", mock.Anything).Return(&types.User{}, nil)
	mockConfig.On("SaveNamespace", mock.Anything).Return(nil)
	mockConfig.On("SaveInsecureSkipTLSVerify", mock.Anything).Return(nil)
	mockConfig.On("SaveTrustedCAFile", mock.Anything).Return(nil)
	mockConfig.On("Timeout").Return(time.Second * 15)

	// We need to call the "configure" command via the rootCmd so the global flags
	// are set
	rootCmd := root.Command()
	cmd := Command(cli)
	require.NoError(t, cmd.Flags().Set("non-interactive", "true"))
	require.NoError(t, cmd.Flags().Set("password", "my-password"))
	require.NoError(t, cmd.Flags().Set("username", "my-user"))
	require.NoError(t, cmd.Flags().Set("url", "http://127.0.0.1:8080"))
	rootCmd.AddCommand(cmd)

	buf := new(bytes.Buffer)
	rootCmd.SetOutput(buf)
	rootCmd.SetArgs([]string{"configure"})
	_, err := rootCmd.ExecuteC()
	out := buf.String()

	assert.NoError(err)
	assert.Empty(out)

	// Make sure the TLS configuration has been saved
	mockConfig.AssertCalled(t, "SaveInsecureSkipTLSVerify", false)
	mockConfig.AssertCalled(t, "SaveTrustedCAFile", "")
}
