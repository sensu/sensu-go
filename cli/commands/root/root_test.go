package root

import (
	"os"
	"testing"

	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/stretchr/testify/assert"
)

func TestCommand(t *testing.T) {
	assert := assert.New(t)
	stdout := test.NewFileCapture(&os.Stdout)

	// Run command w/o any flags
	stdout.Start()
	cmd := Command()

	assert.NotNil(cmd, "Returns a Command instance")
	assert.Equal("sensuctl", cmd.Use, "Configures the name")

	cmd.Run(cmd, []string{})
	stdout.Stop()
	assert.Regexp("Usage:", stdout.Output())
}
