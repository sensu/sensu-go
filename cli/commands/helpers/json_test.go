package helpers

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrintJSON(t *testing.T) {
	assert := assert.New(t)

	testInput := map[string]string{
		"commandLessThan": "echo foo >> output.txt",
		"commandAnd":      "echo bar && exit 1",
	}

	buf := new(bytes.Buffer)
	writer := io.Writer(buf)
	require.NoError(t, PrintJSON(testInput, writer))
	assert.Equal("{\n  \"commandAnd\": \"echo bar && exit 1\",\n  \"commandLessThan\": \"echo foo >> output.txt\"\n}\n\n", buf.String())
}
