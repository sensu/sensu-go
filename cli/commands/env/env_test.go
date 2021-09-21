package env

import (
	"testing"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	mockclient "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/stretchr/testify/assert"
)

func testSetupMocks(t *testing.T, config *mockclient.MockConfig) {
	t.Helper()

	config.On("APIUrl").Return("http://127.0.0.1:8080")
	config.On("Format").Return("none")
	config.On("Tokens").Return(
		corev2.FixtureTokens("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1NjIxODkzMTcsImp0aSI6IjAwZDFlYTE2OGU1MTQ1ZGEzN2U2Njg0YmRlOTgwNDM4Iiwic3ViIjoiYWRtaW4iLCJncm91cHMiOlsiY2x1c3Rlci1hZG1pbnMiLCJzeXN0ZW06dXNlcnMiXSwicHJvdmlkZXIiOnsicHJvdmlkZXJfaWQiOiJiYXNpYyIsInVzZXJfaWQiOiJhZG1pbiJ9fQ.ksuMGCJtkN5724CQ7e2W1P7T2ZPpR8IxU3fH9WhBMLk", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJqdGkiOiI0MGVhYTRiMzRkMzU4YTkzNTY5YzIzZWM1YjcxNmZiMiIsInN1YiI6ImFkbWluIiwiZ3JvdXBzIjpudWxsLCJwcm92aWRlciI6eyJwcm92aWRlcl9pZCI6IiIsInVzZXJfaWQiOiIifX0.7t0qoBvKEkHD1DJbhP-VfSj95yhsFyrPoeFhqEbKOn8"),
	)
	config.On("APIKey").Return("some-api-key")
	config.On("TrustedCAFile").Return("")
	config.On("InsecureSkipTLSVerify").Return(false)
	config.On("Timeout").Return(time.Duration(42) * time.Second)
}

func TestEnvCommandBash(t *testing.T) {
	cli := test.NewMockCLI()
	testSetupMocks(t, cli.Config.(*mockclient.MockConfig))

	cmd := Command(cli)
	_ = cmd.Flags().Set("shell", "bash")
	out, err := test.RunCmd(cmd, nil)
	assert.NoError(t, err)
	assert.Regexp(t, `export SENSU_API_URL="http://127.0.0.1:8080"`, out)
	assert.Regexp(t, `export SENSU_API_KEY="some-api-key"`, out)
	assert.Regexp(t, `export SENSU_TIMEOUT="42s"`, out)
}

func TestEnvCommandCmd(t *testing.T) {
	cli := test.NewMockCLI()
	testSetupMocks(t, cli.Config.(*mockclient.MockConfig))

	cmd := Command(cli)
	_ = cmd.Flags().Set("shell", "cmd")
	out, err := test.RunCmd(cmd, nil)
	assert.NoError(t, err)
	assert.Regexp(t, `SET SENSU_API_URL=http://127.0.0.1:8080`, out)
	assert.Regexp(t, `SET SENSU_API_KEY=some-api-key`, out)
	assert.Regexp(t, `SET SENSU_TIMEOUT=42s`, out)
}

func TestEnvCommandPowershell(t *testing.T) {
	cli := test.NewMockCLI()
	testSetupMocks(t, cli.Config.(*mockclient.MockConfig))

	cmd := Command(cli)
	_ = cmd.Flags().Set("shell", "powershell")
	out, err := test.RunCmd(cmd, nil)
	assert.NoError(t, err)
	assert.Regexp(t, `\$Env:SENSU_API_URL = "http://127.0.0.1:8080"`, out)
	assert.Regexp(t, `\$Env:SENSU_API_KEY = "some-api-key"`, out)
	assert.Regexp(t, `\$Env:SENSU_TIMEOUT = "42s"`, out)
}
