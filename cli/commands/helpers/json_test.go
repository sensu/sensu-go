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
	writer := io.Writer(buf)
	require.NoError(t, PrintJSON(testInput, writer))
	assert.Equal("{\n  \"commandAnd\": \"echo bar && exit 1\",\n  \"commandLessThan\": \"echo foo >> output.txt\"\n}\n\n", buf.String())
}

func TestPrintWrappedJSON(t *testing.T) {
	assert := assert.New(t)

	check := types.FixtureCheckConfig("check")
	check.Command = "echo foo >> output.txt"

	w := types.Wrapper{
		Type:  "CheckConfig",
		Value: check,
	}
	output, err := json.Marshal(w)
	assert.NoError(err)

	buf := new(bytes.Buffer)
	writer := io.Writer(buf)
	require.NoError(t, PrintWrappedJSON(check, writer))
	assert.JSONEq(string(output), buf.String())
}

func TestPrintWrappedJSONList(t *testing.T) {
	assert := assert.New(t)

	check1 := types.FixtureCheckConfig("check1")
	check2 := types.FixtureCheckConfig("check2")

	w1 := types.Wrapper{
		Type:  "CheckConfig",
		Value: check1,
	}
	w2 := types.Wrapper{
		Type:  "CheckConfig",
		Value: check2,
	}
	output1, err := json.Marshal(w1)
	assert.NoError(err)
	output2, err := json.Marshal(w2)
	assert.NoError(err)

	buf := new(bytes.Buffer)
	writer := io.Writer(buf)

	require.NoError(t, PrintWrappedJSONList([]types.Resource{check1, check2}, writer))
	// trim \n and white space for equal comparison
	output3 := strings.Replace(buf.String(), "\n", "", -1)
	output3 = strings.Replace(output3, " ", "", -1)
	assert.Equal(string(output1)+string(output2), output3)
}

func TestPrintFormatted(t *testing.T) {
	assert := assert.New(t)

	check := types.FixtureCheckConfig("check")

	w := types.Wrapper{
		Type:  "CheckConfig",
		Value: check,
	}

	// test wrapped-json format
	output, err := json.Marshal(w)
	assert.NoError(err)

	buf := new(bytes.Buffer)
	writer := io.Writer(buf)

	require.NoError(t, PrintFormatted("", config.FormatWrappedJSON, check, writer, printToList))
	assert.JSONEq(string(output), buf.String())

	// test json format
	output, err = json.Marshal(check)
	assert.NoError(err)

	buf = new(bytes.Buffer)
	writer = io.Writer(buf)

	require.NoError(t, PrintFormatted("", config.FormatJSON, check, writer, printToList))
	assert.JSONEq(string(output), buf.String())

	// test tabular format
	buf = new(bytes.Buffer)
	writer = io.Writer(buf)

	require.NoError(t, PrintFormatted("", config.FormatTabular, check, writer, printToList))
	assert.Equal("=== \n", buf.String()) // empty table

	// test default format
	buf = new(bytes.Buffer)
	writer = io.Writer(buf)

	require.NoError(t, PrintFormatted("none", config.DefaultFormat, check, writer, printToList))
	assert.Equal("=== \n", buf.String()) // empty table

	// test flag override (json format)
	output, err = json.Marshal(check)
	assert.NoError(err)

	buf = new(bytes.Buffer)
	writer = io.Writer(buf)

	require.NoError(t, PrintFormatted(config.FormatJSON, config.FormatWrappedJSON, check, writer, printToList))
	assert.JSONEq(string(output), buf.String())
}

func printToList(v interface{}, writer io.Writer) error {
	cfg := &list.Config{}
	return list.Print(writer, cfg)
}
