package helpers

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/sensu/sensu-go/cli/client/config"
	"github.com/sensu/sensu-go/cli/elements/list"
	"github.com/sensu/sensu-go/types"
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
	require.NoError(t, PrintJSON(testInput, buf))
	assert.Equal("{\n  \"commandAnd\": \"echo bar && exit 1\",\n  \"commandLessThan\": \"echo foo >> output.txt\"\n}\n\n", buf.String())
}

func TestPrintWrappedJSON(t *testing.T) {
	assert := assert.New(t)

	check := types.FixtureCheckConfig("check")
	check.Command = "echo foo >> output.txt"

	w := wrapResource(check)
	output, err := json.Marshal(w)
	assert.NoError(err)

	buf := new(bytes.Buffer)
	require.NoError(t, PrintWrappedJSON(check, buf))
	assert.JSONEq(string(output), buf.String())
}

func TestPrintWrappedJSONList(t *testing.T) {
	assert := assert.New(t)

	check1 := types.FixtureCheckConfig("check1")
	check2 := types.FixtureCheckConfig("check2")

	w1 := wrapResource(check1)
	w2 := wrapResource(check2)

	output1, err := json.Marshal(w1)
	assert.NoError(err)
	output2, err := json.Marshal(w2)
	assert.NoError(err)

	buf := new(bytes.Buffer)

	require.NoError(t, PrintWrappedJSONList([]types.Resource{check1, check2}, buf))
	// compare each string individually
	output3 := strings.Split(buf.String(), "}\n{")
	assert.JSONEq(string(output1), output3[0]+"}")
	assert.JSONEq(string(output2), "{"+output3[1])
}

func TestPrintFormatted(t *testing.T) {
	assert := assert.New(t)

	check := types.FixtureCheckConfig("check")

	w := wrapResource(check)

	// test wrapped-json format
	output, err := json.Marshal(w)
	assert.NoError(err)

	buf := new(bytes.Buffer)

	require.NoError(t, PrintFormatted("", config.FormatWrappedJSON, check, buf, printToList))
	assert.JSONEq(string(output), buf.String())

	// test json format
	output, err = json.Marshal(check)
	assert.NoError(err)

	buf = new(bytes.Buffer)

	require.NoError(t, PrintFormatted("", config.FormatJSON, check, buf, printToList))
	assert.JSONEq(string(output), buf.String())

	// test tabular format
	buf = new(bytes.Buffer)

	require.NoError(t, PrintFormatted("", config.FormatTabular, check, buf, printToList))
	assert.Equal("=== \n", buf.String()) // empty table

	// test default format
	buf = new(bytes.Buffer)

	require.NoError(t, PrintFormatted("none", config.DefaultFormat, check, buf, printToList))
	assert.Equal("=== \n", buf.String()) // empty table

	// test flag override (json format)
	output, err = json.Marshal(check)
	assert.NoError(err)

	buf = new(bytes.Buffer)

	require.NoError(t, PrintFormatted(config.FormatJSON, config.FormatWrappedJSON, check, buf, printToList))
	assert.JSONEq(string(output), buf.String())
}

func printToList(v interface{}, writer io.Writer) error {
	cfg := &list.Config{}
	return list.Print(writer, cfg)
}
