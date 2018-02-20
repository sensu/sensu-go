package importer

import (
	"os"
	"testing"

	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/stretchr/testify/assert"
)

func TestImportCommand(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := ImportCommand(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("import", cmd.Use)
	assert.Regexp("resources", cmd.Short)
}

func TestImportCommandRun(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := ImportCommand(cli)

	out, err := test.RunCmd(cmd, []string{})
	assert.NoError(err)
	assert.Contains(out, "Usage")
}

func TestImportCommandRunWithBadJSON(t *testing.T) {
	assert := assert.New(t)

	reader, writer, _ := os.Pipe()
	_, _ = writer.Write([]byte("one two three"))

	cli := test.NewMockCLI()
	cli.InFile = reader
	cmd := ImportCommand(cli)

	out, err := test.RunCmd(cmd, []string{"in"})
	// Print help usage
	assert.Error(err)
	assert.NotEmpty(out)
}

func TestImportCommandRunWithGoodJSON(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewMockCLI()

	reader, writer, _ := os.Pipe()
	cli.InFile = reader
	_, _ = writer.Write([]byte(`{"ok":"yup"}`))

	cmd := ImportCommand(cli)
	out, err := test.RunCmd(cmd, []string{})

	assert.Error(err)
	assert.Contains(out, "Only importing of legacy settings are supported")
}

func TestImportCommandRunWithLegacyImporter(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewMockCLI()

	reader, writer, _ := os.Pipe()
	cli.InFile = reader

	_, _ = writer.Write([]byte(`{"ok":"yup"}`))

	cmd := ImportCommand(cli)
	_ = cmd.Flags().Set("legacy", "t")
	_ = cmd.Flags().Set("verbose", "t")

	out, err := test.RunCmd(cmd, []string{})

	assert.NotEmpty(out)
	assert.Contains(out, "SUCCESS")
	assert.NoError(err)
}

func TestImportCommandRunWithLegacyImporterInvalidResources(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewMockCLI()

	reader, writer, _ := os.Pipe()
	cli.InFile = reader

	_, _ = writer.Write([]byte(`{"checks": {"frank%%%": {}}}`))

	cmd := ImportCommand(cli)
	_ = cmd.Flags().Set("legacy", "t")
	_ = cmd.Flags().Set("verbose", "t")

	out, err := test.RunCmd(cmd, []string{})

	assert.NotEmpty(out)
	assert.Contains(out, "ERROR")
	assert.Error(err)
}
