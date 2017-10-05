package helpers

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrintJSON(t *testing.T) {
	assert := assert.New(t)

	testInput := map[string]string{
		"commandLessThan": "echo foo >> output.txt",
		"commandAnd":      "echo bar && exit 1",
	}

	buf := new(bytes.Buffer)
	writer := io.Writer(buf)
	if err := PrintJSON(testInput, writer); err != nil {
		assert.FailNow("failed to parse JSON due to error %s", err)
	}
	assert.Equal(string(buf.Bytes()), "{\n  \"commandAnd\": \"echo bar && exit 1\",\n  \"commandLessThan\": \"echo foo >> output.txt\"\n}\n\n")
}
