package user

import (
	"testing"

	//	clientmock "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/stretchr/testify/assert"
	//	"github.com/stretchr/testify/mock"
	//	"github.com/stretchr/testify/require"
)

func TestTestCredsCommand(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := TestCredsCommand(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("test-creds", cmd.Use)
	assert.Regexp("test user credentials", cmd.Short)
}
