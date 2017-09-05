package table

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStandardTable(t *testing.T) {
	assert := assert.New(t)
	writer := exWriter{}

	table := New([]*Column{
		{
			Title:       "One",
			ColumnStyle: PrimaryTextStyle, // Make each row blue
			CellTransformer: func(_ interface{}) string {
				return "cell-one"
			},
		},
		{
			Title: "Two",
			CellTransformer: func(_ interface{}) string {
				return "cell-two"
			},
		},
	})
	table.Render(
		&writer,
		[]*Row{
			{Value: "blah"},
			{Value: "blah blah"},
		},
	)

	lines := strings.Split(writer.result, "\n")
	heading := lines[0]
	row1 := lines[2]
	row2 := lines[3]

	// Four lines of output (headings, separator, row1, row2, & new line)
	assert.Len(lines, 5)

	// Ensure that both headings for each column are present and styled
	assert.Contains(heading, "One")
	assert.Contains(heading, TitleStyle("One"))
	assert.Contains(heading, "Two")
	assert.Contains(heading, TitleStyle("Two"))

	// Ensure that the both columns are present and styled appropriately
	assert.Contains(row1, "cell-one")
	assert.Contains(row1, PrimaryTextStyle("cell-one"))
	assert.Contains(row1, "cell-two")
	assert.NotContains(row1, PrimaryTextStyle("cell-two"))

	// Ensure that the both columns are present and styled appropriately
	assert.Contains(row2, "cell-one")
	assert.Contains(row2, PrimaryTextStyle("cell-one"))
	assert.Contains(row2, "cell-two")
	assert.NotContains(row2, PrimaryTextStyle("cell-two"))
}

type exWriter struct {
	result string
}

func (w *exWriter) Clean() {
	w.result = ""
}

func (w *exWriter) Write(p []byte) (int, error) {
	w.result += string(p)
	return 0, nil
}
