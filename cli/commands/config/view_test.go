package config

import (
	"testing"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	clienttest "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/stretchr/testify/assert"
)

func TestViewCommand(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := ViewCommand(cli)

	config := cli.Config.(*clienttest.MockConfig)
	config.On("APIUrl").Return("http://127.0.0.1:8080")
	config.On("Format").Return("none")

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("view", cmd.Use)
	assert.Regexp("Display active configuration", cmd.Short)
}

func TestViewExec(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := ViewCommand(cli)

	config := cli.Config.(*clienttest.MockConfig)
	config.On("APIUrl").Return("http://127.0.0.1:8080")
	config.On("Format").Return("none")
	config.On("Tokens").Return(
		corev2.FixtureTokens("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1NjIxODkzMTcsImp0aSI6IjAwZDFlYTE2OGU1MTQ1ZGEzN2U2Njg0YmRlOTgwNDM4Iiwic3ViIjoiYWRtaW4iLCJncm91cHMiOlsiY2x1c3Rlci1hZG1pbnMiLCJzeXN0ZW06dXNlcnMiXSwicHJvdmlkZXIiOnsicHJvdmlkZXJfaWQiOiJiYXNpYyIsInVzZXJfaWQiOiJhZG1pbiJ9fQ.ksuMGCJtkN5724CQ7e2W1P7T2ZPpR8IxU3fH9WhBMLk", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJqdGkiOiI0MGVhYTRiMzRkMzU4YTkzNTY5YzIzZWM1YjcxNmZiMiIsInN1YiI6ImFkbWluIiwiZ3JvdXBzIjpudWxsLCJwcm92aWRlciI6eyJwcm92aWRlcl9pZCI6IiIsInVzZXJfaWQiOiIifX0.7t0qoBvKEkHD1DJbhP-VfSj95yhsFyrPoeFhqEbKOn8"),
	)
	config.On("Timeout").Return(time.Second * 15)

	out, err := test.RunCmd(cmd, nil)
	assert.Regexp("Active Configuration", out)
	assert.Regexp("admin", out)
	assert.Nil(err, "Should not produce any errors")
}

func TestViewErr(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := ViewCommand(cli)

	config := cli.Config.(*clienttest.MockConfig)
	config.On("APIUrl").Return("")
	config.On("Format").Return("")
	config.On("Tokens").Return((*corev2.Tokens)(nil))

	_, err := test.RunCmd(cmd, nil)
	assert.NotNil(err, "Should not produce any errors")
}
