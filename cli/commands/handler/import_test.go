package handler

import (
	"encoding/json"
	"errors"
	"os"
	"testing"

	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestImportCommand(t *testing.T) {
	assert := assert.New(t)

	cli := newCLI()
	cmd := ImportCommand(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("import", cmd.Use)
	assert.Regexp("handler", cmd.Short)
}

func TestImportCommandRunE(t *testing.T) {
	assert := assert.New(t)

	cli := newCLI()
	cmd := ImportCommand(cli)

	out, err := test.RunCmd(cmd, []string{"in"})
	assert.NoError(err)
	assert.Contains(out, "Usage")
}

func TestImportCommandRunEWithBadJSON(t *testing.T) {
	assert := assert.New(t)

	pipeWithContents([]byte("one two  {three: 123}"), func(reader *os.File) {
		cli := newCLI()
		cli.InFile = reader
		cmd := ImportCommand(cli)

		out, err := test.RunCmd(cmd, []string{"in"})
		assert.Error(err)
		assert.Empty(out)
	})
}

func TestImportCommandRunEWithGoodJSON(t *testing.T) {
	assert := assert.New(t)

	handler := types.FixtureHandler("foo")
	handlerBytes, _ := json.Marshal(handler)

	pipeWithContents(handlerBytes, func(reader *os.File) {
		cli := newCLI()
		cli.InFile = reader

		client := cli.Client.(*client.MockClient)
		client.On("CreateHandler", mock.Anything).Return(nil)

		cmd := ImportCommand(cli)
		out, err := test.RunCmd(cmd, []string{})

		assert.NoError(err)
		assert.Contains(out, "Imported")
	})
}

func TestImportCommandRunEWithBadResponse(t *testing.T) {
	assert := assert.New(t)
	cli := newCLI()

	handler := types.FixtureHandler("foo")
	handlerBytes, _ := json.Marshal(handler)

	pipeWithContents(handlerBytes, func(reader *os.File) {
		cli.InFile = reader

		client := cli.Client.(*client.MockClient)
		client.On("CreateHandler", mock.Anything).Return(errors.New("a"))

		cmd := ImportCommand(cli)
		out, err := test.RunCmd(cmd, []string{})

		assert.NotContains(out, "Imported")
		assert.Empty(out)
		assert.Error(err)
	})
}

// NOTE: Windows issues with pipes seems to mean that writing and reading a pipe
// in the same goroutine can cause panics
func pipeWithContents(c []byte, fn func(*os.File)) {
	reader, writer, _ := os.Pipe()
	writer.Write(c)

	ch := make(chan struct{})
	go func() {
		fn(reader)
		close(ch)
	}()
	<-ch
}
